package main

import (
	"context"
	"flag"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/revrost/counterspell/internal/config"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/handlers"
	"github.com/revrost/counterspell/internal/services"
	"github.com/revrost/counterspell/internal/tunnel"
	"github.com/revrost/counterspell/ui"
)

// Context key for subdomain
type contextKey string

const subdomainKey contextKey = "subdomain"

func main() {
	// Parse flags
	addr := flag.String("addr", ":8710", "Server address")
	flag.Parse()

	// Setup log output: always write to both stdout and server.log
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		os.Exit(1)
	}
	defer func() { _ = logFile.Close() }()
	logOutput := io.MultiWriter(os.Stdout, logFile)

	// Setup logger
	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting server",
		"database_path", cfg.DatabasePath,
		"data_dir", cfg.DataDir,
		"worker_pool_size", cfg.WorkerPoolSize,
		"max_tasks_per_user", cfg.MaxTasksPerUser,
	)

	// Ensure directory structure
	if err := config.EnsureDirectories(cfg); err != nil {
		logger.Error("Failed to create directories", "error", err)
		os.Exit(1)
	}

	// Connect to SQLite
	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(ctx); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Ensure auth + machine identity before starting server (Handshake starts here)
	authService := services.NewOAuthService(database, cfg)
	authResult, err := authService.EnsureAuthenticated(ctx)
	if err != nil {
		logger.Error("Authentication failed. Exiting application", "error", err)
		os.Exit(1)
	}
	logger.Info("Authenticated", "subdomain", authResult.Subdomain, "machine_id", authResult.MachineID)

	// Create event bus
	eventBus := services.NewEventBus()

	// Start session syncer (imports existing CLI sessions and tails for updates)
	repo := services.NewRepository(database)
	syncCtx, syncCancel := context.WithCancel(ctx)
	syncer := services.NewSessionSyncer(repo, eventBus)
	syncer.Start(syncCtx)

	// Create handlers with shared database
	h, err := handlers.NewHandlers(database, eventBus, cfg)
	if err != nil {
		logger.Error("Failed to create handlers", "error", err)
		os.Exit(1)
	}

	// Setup router
	slog.Info("Setting up router")
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Subdomain extraction middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			if idx := strings.Index(host, ":"); idx != -1 {
				host = host[:idx]
			}

			parts := strings.Split(host, ".")
			var subdomain string
			if len(parts) >= 3 && parts[0] != "www" {
				subdomain = parts[0]
			}

			ctx := context.WithValue(r.Context(), subdomainKey, subdomain)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, map[string]string{"status": "ok"})
		})

		// Debug endpoint to show subdomain
		r.Get("/debug/subdomain", func(w http.ResponseWriter, r *http.Request) {
			subdomain := SubdomainFromContext(r.Context())
			render.JSON(w, r, map[string]string{"subdomain": subdomain})
		})

		// UI logging - no auth required so errors can be logged even when auth fails
		// r.Post("/api/v1/log", h.HandleUILog)

		// Auth + session endpoints (needed before auth)
		r.Get("/api/v1/session", h.HandleGetSession)
		r.Get("/api/v1/auth/login", h.HandleAuthLogin)
	})

	// Protected routes (require machine auth)
	r.Group(func(r chi.Router) {
		r.Use(h.RequireMachineAuth)
		// GitHub OAuth routes
		r.Get("/api/v1/github/authorize", h.HandleGitHubLogin)
		r.Get("/api/v1/github/callback", h.HandleGitHubCallback)
		r.Get("/api/v1/github/repos", h.HandleGitHubRepos)

		// Unified SSE endpoint
		r.Get("/api/v1/events", h.HandleSSE)

		// Home page actions, tasks are like inbox
		r.Get("/api/v1/tasks", h.HandleListTask)
		r.Post("/api/v1/tasks", h.HandleAddTask)
		r.Get("/api/v1/task/{id}", h.HandleGetTask)
		r.Get("/api/v1/task/{id}/diff", h.HandleGetTaskDiff)
		r.Get("/api/v1/settings", h.HandleGetSettings)
		r.Get("/api/v1/files/search", h.HandleFileSearch)

		// Settings and transcription
		r.Post("/api/v1/settings", h.HandleSaveSettings)
		r.Post("/api/v1/transcribe", h.HandleTranscribe)

		// Task Actions
		r.Post("/api/v1/tasks/{id}/chat", h.HandleActionChat)
		r.Post("/api/v1/tasks/{id}/clear", h.HandleActionClear)
		r.Post("/api/v1/tasks/{id}/retry", h.HandleActionRetry)
		r.Post("/api/v1/tasks/{id}/merge", h.HandleActionMerge)
		r.Post("/api/v1/tasks/{id}/pr", h.HandleActionPR)
		r.Post("/api/v1/tasks/{id}/discard", h.HandleActionDiscard)

	})

	// Serve Svelte UI (embedded SPA build) - MUST BE LAST
	svelteFS, err := ui.Static()
	if err != nil {
		log.Printf("Failed to load Svelte UI: %v, running without static assets", err)
	} else {
		r.Get("/*", spaHandler(svelteFS))
	}

	// Start server
	server := &http.Server{
		Addr:        *addr,
		Handler:     r,
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", "addr", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Start Cloudflare tunnel (best effort)
	localURL := localURLFromAddr(*addr)
	var tunnelProc *tunnel.CloudflareTunnel
	if authResult.TunnelToken != "" {
		proc, err := tunnel.StartCloudflare(ctx, authResult.TunnelToken, localURL, "", logger)
		if err != nil {
			logger.Warn("Failed to start tunnel", "error", err)
		} else {
			tunnelProc = proc
			logger.Info("Tunnel started", "url", "https://"+authResult.Subdomain+".counterspell.app", "local_url", localURL)
		}
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Stop session syncer
	syncCancel()
	syncer.Shutdown()

	// Shutdown handlers (stops all active orchestrators)
	h.Shutdown()

	// Shutdown event bus (stops cleanup goroutine)
	eventBus.Shutdown()

	if tunnelProc != nil {
		_ = tunnelProc.Stop()
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}

func localURLFromAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0") {
		parts := strings.Split(addr, ":")
		if len(parts) == 2 && parts[1] != "" {
			return "http://localhost:" + parts[1]
		}
	}
	if strings.HasPrefix(addr, "127.0.0.1") || strings.HasPrefix(addr, "localhost") {
		return "http://" + addr
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}

// SubdomainFromContext extracts the subdomain from the request context.
// Returns empty string if no subdomain is present.
func SubdomainFromContext(ctx context.Context) string {
	if subdomain, ok := ctx.Value("subdomain").(string); ok {
		return subdomain
	}
	return ""
}

// spaHandler returns an http.Handler that serves the SPA with client-side routing.
// If the requested file exists, it serves it directly. Otherwise, it serves index.html
// to enable client-side routing.
func spaHandler(svelteFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(svelteFS))
	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		f, err := svelteFS.Open(cleanPath)
		if err == nil {
			_ = f.Close()
		} else if os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	}
}
