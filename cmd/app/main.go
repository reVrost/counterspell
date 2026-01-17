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

	// Log startup mode
	if cfg.MultiTenant {
		logger.Info("Starting in MULTI-TENANT mode",
			"supabase_url", cfg.SupabaseURL,
			"worker_pool_size", cfg.WorkerPoolSize,
			"max_tasks_per_user", cfg.MaxTasksPerUser,
		)
	} else {
		logger.Info("Starting in SINGLE-PLAYER mode",
			"data_dir", cfg.DataDir,
		)
	}

	// Ensure directory structure
	if err := config.EnsureDirectories(cfg); err != nil {
		logger.Error("Failed to create directories", "error", err)
		os.Exit(1)
	}

	// In single-player mode, migrate from old structure
	if !cfg.MultiTenant {
		if err := config.MigrateFromSingleUser(cfg); err != nil {
			logger.Warn("Migration from single-user failed (may be first run)", "error", err)
		}
	}

	// Create DB manager (handles per-user databases)
	dbManager := db.NewDBManager(cfg)
	userRegistry := services.NewUserManagerRegistry(cfg, dbManager)
	eventBus := services.NewEventBus()

	// Reset stuck tasks on startup for default user (single-player mode)
	if !cfg.MultiTenant {
		defaultDB, err := dbManager.GetDB("default")
		if err != nil {
			logger.Error("Failed to open default database", "error", err)
			os.Exit(1)
		}
		taskSvc := services.NewTaskService(defaultDB)
		ctx := context.Background()
		if err := taskSvc.ResetInProgress(ctx); err != nil {
			logger.Error("Failed to reset in-progress tasks", "error", err)
		}
	}

	// Create auth middleware
	authMiddleware := auth.NewMiddleware(cfg, dbManager)

	// Create handlers with user registry for per-user service creation
	h, err := handlers.NewHandlers(userRegistry, eventBus, cfg)
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

	// Protected routes (auth required in multi-tenant mode)
	r.Group(func(r chi.Router) {
		// Apply auth middleware - in single-player mode this sets userID to "default"
		r.Use(authMiddleware.RequireAuth)

		// Main app routes (authenticated users go to /app or /feed)
		r.Get("/app", h.RenderApp)
		r.Get("/feed", h.HandleFeed)
		r.Get("/task/{id}", h.HandleTaskDetailUI)

		// Unified SSE endpoint
		r.Get("/events", h.HandleSSE)

		// Actions
		r.Post("/add-task", h.HandleAddTask)
		r.Post("/action/retry/{id}", h.HandleActionRetry)
		r.Post("/action/clear/{id}", h.HandleActionClear)
		r.Post("/action/merge/{id}", h.HandleActionMerge)
		r.Post("/action/pr/{id}", h.HandleActionPR)
		r.Post("/action/discard/{id}", h.HandleActionDiscard)
		r.Post("/action/chat/{id}", h.HandleActionChat)

		// Merge
		r.Post("/action/resolve-conflict/{id}", h.HandleResolveConflict)
		r.Post("/action/abort-merge/{id}", h.HandleAbortMerge)
		r.Post("/action/complete-merge/{id}", h.HandleCompleteMerge)

		// Settings and transcription
		r.Post("/settings", h.HandleSaveSettings)
		r.Post("/transcribe", h.HandleTranscribe)
		r.Get("/api/files/search", h.HandleFileSearch)

		// Auth management
		r.Post("/auth/logout", h.HandleLogout)

		// GitHub OAuth
		r.Get("/github/authorize", h.HandleGitHubAuthorize)
		r.Post("/disconnect", h.HandleDisconnect)
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

	userRegistry.Shutdown()
	if err := dbManager.CloseAll(); err != nil {
		logger.Error("Failed to close databases", "error", err)
	}

	logger.Info("Server stopped")
}
