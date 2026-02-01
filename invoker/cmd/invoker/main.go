package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/revrost/invoker/internal/auth"
	"github.com/revrost/invoker/internal/config"
	"github.com/revrost/invoker/internal/db"
	"github.com/revrost/invoker/internal/fly"
	"github.com/revrost/invoker/ui"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	ctx := context.Background()

	// Initialize database connection
	var dbConn *db.DB
	var database db.Repository
	if cfg.DatabaseURL != "" {
		var err error
		const maxAttempts = 10
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			dbConn, err = db.NewDB(cfg.DatabaseURL)
			if err == nil {
				break
			}
			log.Printf("Failed to connect to database (attempt %d/%d): %v", attempt, maxAttempts, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		if err != nil {
			log.Printf("Failed to connect to database after %d attempts: %v", maxAttempts, err)
			os.Exit(1)
		}
		defer dbConn.Close()
		database = db.NewPostgresRepository(dbConn)
		log.Println("Database connected")
	} else {
		log.Println("DATABASE_URL not set, running without database")
		database = nil
	}

	// Initialize Supabase auth
	var supabaseAuth *auth.SupabaseAuth
	if cfg.SupabaseURL != "" && cfg.SupabaseAnonKey != "" {
		var err error
		supabaseAuth, err = auth.NewSupabaseAuth(cfg.SupabaseURL, cfg.SupabaseAnonKey, cfg.JWTSecret)
		if err != nil {
			log.Printf("Failed to initialize Supabase auth: %v, running without auth", err)
			supabaseAuth = nil
		} else {
			log.Println("Supabase auth initialized")
		}
	} else {
		log.Println("SUPABASE_URL or SUPABASE_ANON_KEY not set, running without Supabase auth")
	}

	// Initialize Fly.io client
	var flyClient *fly.Client
	var flyService *fly.Service
	if cfg.FlyAPIToken != "" && cfg.FlyOrg != "" {
		flyClient = fly.NewClient(cfg.FlyAPIToken, cfg.FlyOrg)
		if database != nil {
			flyService = fly.NewService(flyClient, database, cfg.FlyAppName, cfg.FlyDockerImage, cfg.FlyRegion)
			log.Printf("Fly.io service initialized (app=%s, region=%s)", cfg.FlyAppName, cfg.FlyRegion)
		}
	} else {
		log.Println("FLY_API_TOKEN or FLY_ORG not set, running without Fly.io integration")
	}

	// Initialize handlers
	authHandler := auth.NewHandler(supabaseAuth, database, flyService)
	oauthHandler := auth.NewOAuthHandler(supabaseAuth, database, cfg)
	machineHandler := auth.NewMachineHandler(database, cfg)
	_ = auth.NewEdgeMiddleware(supabaseAuth, database) // For edge/proxy use

	// Example of protected route (uncomment to use)
	// r.With(authHandler.JWTMiddleware).Get("/api/protected", handleProtected)

	// Run migrations (synchronously to avoid races in local/dev)
	if dbConn != nil {
		if err := db.RunMigrations(ctx, dbConn.Pool); err != nil {
			slog.Error("Failed to run migrations, continuing with degraded database functionality", "error", err)
		}
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/health", handleHealth)
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		handleReady(w, r, dbConn)
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/profiles", authHandler.SyncProfile)
			r.With(authHandler.JWTMiddleware).Get("/profile", authHandler.GetProfile)
			// OAuth flow handlers
			r.Post("/url", oauthHandler.CreateCLIAuthURL)
			r.Post("/poll", oauthHandler.PollOAuthCode)
			r.Post("/exchange", oauthHandler.ExchangeOAuthCode)
			// Device flow handlers
		})

		// OAuth callback (browser-facing, outside /api/v1)
		r.Get("/auth/callback", oauthHandler.OAuthCallback)

		// VM management routes
		r.Route("/vm", func(r chi.Router) {
			r.Post("/start", handleVMStart)
			r.Get("/status", handleVMStatus)
			r.Delete("/stop", handleVMStop)
		})

		// Machine registry routes
		r.Route("/machines", func(r chi.Router) {
			r.With(authHandler.JWTMiddleware).Get("/", machineHandler.ListMachines)
			r.With(authHandler.JWTMiddleware).Get("/{id}", machineHandler.GetMachine)
			// Machine management (requires machine JWT)
			r.With(auth.RequireMachineJWT(cfg.JWTSecret)).Post("/register", machineHandler.RegisterMachine)
			r.With(authHandler.JWTMiddleware).Post("/{machine_id}/revoke", machineHandler.RevokeMachine)
		})

		// Routing table routes
		r.Get("/routing/{subdomain}", handleGetRouting)

		// Waitlist route
		r.Post("/waitlist", func(w http.ResponseWriter, r *http.Request) {
			handleJoinWaitlist(w, r, dbConn)
		})
	})

	// Serve Svelte UI (embedded build) - MUST BE LAST
	svelteFS, err := ui.Static()
	if err != nil {
		log.Printf("Failed to load Svelte UI: %v, running without static assets", err)
	} else {
		fileServer := http.FileServer(http.FS(svelteFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Normalize path and try to open file
			cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
			f, err := svelteFS.Open(cleanPath)
			if err == nil {
				f.Close()
			} else if os.IsNotExist(err) {
				// File doesn't exist - serve index.html for SPA routing
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting invoker on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"version": os.Getenv("APP_VERSION"),
	}); err != nil {
		log.Printf("Failed to encode health response: %v", err)
	}
}

func handleReady(w http.ResponseWriter, r *http.Request, database *db.DB) {
	dbConnected := database != nil
	if database != nil {
		ctx := r.Context()
		if err := database.Pool.Ping(ctx); err != nil {
			dbConnected = false
			log.Printf("Database ping failed: %v", err)
		}
	}

	status := http.StatusOK
	if !dbConnected {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"status": func() string {
			if dbConnected {
				return "ok"
			} else {
				return "degraded"
			}
		}(),
		"database": dbConnected,
	}); err != nil {
		log.Printf("Failed to encode ready response: %v", err)
	}
}

// Placeholder handlers - will be implemented in future tasks

func handleVMStart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not implemented yet"}); err != nil {
		log.Printf("Failed to encode VM start response: %v", err)
	}
}

func handleVMStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not implemented yet"}); err != nil {
		log.Printf("Failed to encode VM status response: %v", err)
	}
}

func handleVMStop(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not implemented yet"}); err != nil {
		log.Printf("Failed to encode VM stop response: %v", err)
	}
}

// Legacy placeholders removed; machine routes handled by MachineHandler.

func handleGetRouting(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": "Not implemented yet"}); err != nil {
		log.Printf("Failed to encode get routing response: %v", err)
	}
}

func handleJoinWaitlist(w http.ResponseWriter, r *http.Request, dbConn *db.DB) {
	if dbConn == nil {
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err := dbConn.Pool.Exec(ctx, "INSERT INTO waitlist (email) VALUES ($1) ON CONFLICT (email) DO NOTHING", req.Email)
	if err != nil {
		log.Printf("Failed to add to waitlist: %v", err)
		http.Error(w, "Failed to join waitlist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to the hunt"})
}
