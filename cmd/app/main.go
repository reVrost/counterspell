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
	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/handlers"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/ui"
)

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

	// Create event bus
	eventBus := services.NewEventBus()

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

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, map[string]string{"status": "ok"})
		})

		// UI logging - no auth required so errors can be logged even when auth fails
		// r.Post("/api/v1/log", h.HandleUILog)
	})

	// Protected routes (auth not required for local-first)
	r.Group(func(r chi.Router) {
		// GitHub OAuth routes
		r.Get("/api/v1/github/authorize", h.HandleGitHubLogin)
		r.Get("/api/v1/github/callback", h.HandleGitHubCallback)
		r.Get("/api/v1/github/repos", h.HandleGitHubRepos)

		// Unified SSE endpoint
		r.Get("/api/v1/events", h.HandleSSE)

		// Home page actions, tasks are like inbox
		r.Get("/api/v1/tasks", h.HandleAPITasks)
		r.Post("/api/v1/tasks", h.HandleAddTask)
		r.Get("/api/v1/task/{id}", h.HandleAPITask)
		r.Get("/api/v1/session", h.HandleAPISession)
		r.Get("/api/v1/settings", h.HandleAPISettings)
		r.Get("/api/v1/files/search", h.HandleFileSearch)

		// Settings and transcription
		r.Post("/api/v1/settings", h.HandleSaveSettings)
		r.Post("/api/v1/transcribe", h.HandleTranscribe)

		// Task Actions
		r.Post("/api/v1/tasks/{id}/clear", h.HandleActionClear)
		r.Post("/api/v1/tasks/{id}/retry", h.HandleActionRetry)
		r.Post("/api/v1/tasks/{id}/merge", h.HandleActionMerge)
		r.Post("/api/v1/tasks/{id}/pr", h.HandleActionPR)
		r.Post("/api/v1/tasks/{id}/discard", h.HandleActionDiscard)

		// Task messages
		r.Get("/api/v1/task/{id}/messages", h.HandleAPIMessages)

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

// spaHandler serves the SvelteKit SPA from an embedded filesystem.
// It serves static assets directly and falls back to index.html for client-side routing.
func spaHandler(fsys fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to serve the file directly
		f, err := fsys.Open(path)
		if err == nil {
			_ = f.Close()
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
		defer func() { _ = indexFile.Close() }()

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
