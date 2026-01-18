package main

import (
	"context"
	"flag"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/handlers"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/ui"
)

func main() {
	// Parse flags
	addr := flag.String("addr", ":8710", "Server address")
	logFile := flag.String("log", "", "Log file path (writes to both stdout and file)")
	flag.Parse()

	// Setup log output (stdout + optional file)
	var logOutput io.Writer = os.Stdout
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			slog.Error("Failed to open log file", "path", *logFile, "error", err)
			os.Exit(1)
		}
		defer f.Close()
		logOutput = io.MultiWriter(os.Stdout, f)
	}

	// Setup logger
	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
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

		// Auth routes (login page, OAuth callbacks)
		r.Get("/api/v1/auth/oauth/{provider}", h.HandleOAuth)
		r.Get("/api/v1/auth/callback", h.HandleAuthCallback)
		r.Get("/api/v1/github/callback", h.HandleGitHubCallback)

		// UI logging - no auth required so errors can be logged even when auth fails
		r.Post("/api/v1/log", h.HandleUILog)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		// Apply auth middleware - sets userID from JWT or defaults to "default"
		r.Use(authMiddleware.RequireAuth)

		// Unified SSE endpoint
		r.Get("/api/v1/events", h.HandleSSE)

		// Actions
		r.Post("/api/v1/add-task", h.HandleAddTask)
		r.Post("/api/v1/action/clear/{id}", h.HandleActionClear)
		r.Post("/api/v1/action/retry/{id}", h.HandleActionRetry)
		r.Post("/api/v1/action/merge/{id}", h.HandleActionMerge)
		r.Post("/api/v1/action/pr/{id}", h.HandleActionPR)
		r.Post("/api/v1/action/discard/{id}", h.HandleActionDiscard)
		r.Post("/api/v1/action/chat/{id}", h.HandleActionChat)

		// Merge
		r.Post("/api/v1/action/resolve-conflict/{id}", h.HandleResolveConflict)
		r.Post("/api/v1/action/abort-merge/{id}", h.HandleAbortMerge)
		r.Post("/api/v1/action/complete-merge/{id}", h.HandleCompleteMerge)

		// JSON API endpoints (for SvelteKit SPA)
		r.Get("/api/v1/tasks", h.HandleAPITasks)
		r.Get("/api/v1/task/{id}", h.HandleAPITask)
		r.Get("/api/v1/session", h.HandleAPISession)
		r.Get("/api/v1/settings", h.HandleAPISettings)
		r.Get("/api/v1/files/search", h.HandleFileSearch)
		r.Get("/api/v1/github/repos", h.HandleGitHubRepos)
		r.Post("/api/v1/github/sync", h.HandleSyncRepos)
		r.Post("/api/v1/project/activate", h.HandleActivateProject)

		// Settings and transcription
		r.Post("/api/v1/settings", h.HandleSaveSettings)
		r.Post("/api/v1/transcribe", h.HandleTranscribe)

		// Logging (for agent debugging)
		r.Get("/api/v1/logs", h.HandleReadLogs)

		// Auth management
		r.Post("/api/v1/logout", h.HandleLogout)
		r.Post("/api/v1/disconnect", h.HandleDisconnect)

		// GitHub OAuth (for users already authenticated via Supabase)
		r.Get("/api/v1/github/authorize", h.HandleGitHubAuthorize)
	})

	// Serve SvelteKit SPA from embedded filesystem
	// This must come after API routes - it's a catch-all for the SPA
	r.Get("/*", spaHandler(ui.DistDirFs))

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

// spaHandler serves the SvelteKit SPA from an embedded filesystem.
// It serves static assets directly and falls back to index.html for client-side routing.
func spaHandler(fsys fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to serve the file directly
		f, err := fsys.Open(path)
		if err == nil {
			f.Close()
			// File exists - serve it with caching for immutable assets
			if strings.HasPrefix(path, "_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		// File not found - serve index.html for SPA routing
		indexFile, err := fsys.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Failed to stat index.html", http.StatusInternalServerError)
			return
		}

		// Serve index.html
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	}
}
