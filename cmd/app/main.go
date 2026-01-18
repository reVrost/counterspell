package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/handlers"
	"github.com/revrost/code/counterspell/internal/services"
)

func main() {
	// Parse flags
	addr := flag.String("addr", ":8710", "Server address")
	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting server",
		"database_url", maskDatabaseURL(cfg.DatabaseURL),
		"data_dir", cfg.DataDir,
		"worker_pool_size", cfg.WorkerPoolSize,
		"max_tasks_per_user", cfg.MaxTasksPerUser,
	)

	// Ensure directory structure
	if err := config.EnsureDirectories(cfg); err != nil {
		logger.Error("Failed to create directories", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabaseURL)
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

	// Create event bus
	eventBus := services.NewEventBus()

	// Create auth middleware
	authMiddleware := auth.NewMiddleware(cfg)

	// Create handlers with shared database
	h, err := handlers.NewHandlers(database, eventBus, cfg)
	if err != nil {
		logger.Error("Failed to create handlers", "error", err)
		os.Exit(1)
	}

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})
		// Static files
		r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
		r.HandleFunc("/static/manifest.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/manifest+json")
			http.ServeFile(w, r, "static/manifest.json")
		})
		r.HandleFunc("/static/sw.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Service-Worker-Allowed", "/")
			http.ServeFile(w, r, "static/sw.js")
		})

		// Landing page (public home)
		r.Get("/", h.HandleHome)
		// Auth routes (login page, OAuth callbacks)
		r.Get("/auth/oauth/{provider}", h.HandleOAuth)
		r.Get("/auth/callback", h.HandleAuthCallback)
		r.Get("/github/callback", h.HandleGitHubCallback)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		// Apply auth middleware - sets userID from JWT or defaults to "default"
		r.Use(authMiddleware.RequireAuth)

		// Unified SSE endpoint
		r.Get("/events", h.HandleSSE)

		// Actions
		r.Post("/api/add-task", h.HandleAddTask)
		r.Post("/api/action/clear/{id}", h.HandleActionClear)
		r.Post("/api/action/retry/{id}", h.HandleActionRetry)
		r.Post("/api/action/merge/{id}", h.HandleActionMerge)
		r.Post("/api/action/pr/{id}", h.HandleActionPR)
		r.Post("/api/action/discard/{id}", h.HandleActionDiscard)
		r.Post("/api/action/chat/{id}", h.HandleActionChat)

		// Merge
		r.Post("/api/action/resolve-conflict/{id}", h.HandleResolveConflict)
		r.Post("/api/action/abort-merge/{id}", h.HandleAbortMerge)
		r.Post("/api/action/complete-merge/{id}", h.HandleCompleteMerge)

		// JSON API endpoints (for SvelteKit SPA)
		r.Get("/api/feed", h.HandleAPIFeed)
		r.Get("/api/task/{id}", h.HandleAPITask)
		r.Get("/api/session", h.HandleAPISession)
		r.Get("/api/settings", h.HandleAPISettings)
		r.Get("/api/files/search", h.HandleFileSearch)
		r.Get("/api/github/repos", h.HandleGitHubRepos)
		r.Post("/api/project/activate", h.HandleActivateProject)

		// Settings and transcription
		r.Post("/api/settings", h.HandleSaveSettings)
		r.Post("/api/transcribe", h.HandleTranscribe)

		// Auth management
		r.Post("/api/logout", h.HandleLogout)
		r.Post("/api/disconnect", h.HandleDisconnect)

		// GitHub OAuth (for users already authenticated via Supabase)
		r.Get("/api/github/authorize", h.HandleGitHubAuthorize)
	})

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

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}

// maskDatabaseURL masks the password in a database URL for logging
func maskDatabaseURL(url string) string {
	if url == "" {
		return "(not set)"
	}
	// Simple masking - just show the host part
	if len(url) > 30 {
		return url[:20] + "..."
	}
	return url[:10] + "..."
}
