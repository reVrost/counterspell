package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/handlers"
	"github.com/revrost/code/counterspell/internal/services"
)

func main() {
	// Parse flags
	addr := flag.String("addr", ":8710", "Server address")
	dbPath := flag.String("db", "./data/pocket-cto.db", "Database path")
	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		logger.Error("Failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.Open(*dbPath)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	if err := database.Close(); err != nil {
		logger.Error("Failed to close database", "error", err)
	}

	// Create services
	taskSvc := services.NewTaskService(database)
	eventBus := services.NewEventBus()

	// Create git worktree manager
	repoPath, _ := os.Getwd()
	worktreeMgr := git.NewWorktreeManager(repoPath)

	// Create agent runner
	agentRunner := services.NewAgentRunner(taskSvc, eventBus, worktreeMgr)

	// Reset stuck tasks on startup
	ctx := context.Background()
	if err := taskSvc.ResetInProgress(ctx); err != nil {
		logger.Error("Failed to reset in-progress tasks", "error", err)
	}

	// Create handlers
	h := handlers.NewHandlers(taskSvc, eventBus, agentRunner, database)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	h.RegisterRoutes(r)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("web/static"))))

	// PWA manifest
	r.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		manifest := `{
			"name": "counterspell",
			"short_name": "PocketCTO",
			"description": "Mobile-first AI agent orchestration",
			"start_url": "/",
			"display": "standalone",
			"background_color": "#000000",
			"theme_color": "#000000",
			"orientation": "portrait",
			"icons": [
				{
					"src": "/static/icon-192.png",
					"sizes": "192x192",
					"type": "image/png"
				},
				{
					"src": "/static/icon-512.png",
					"sizes": "512x512",
					"type": "image/png"
				}
			]
		}`
		if _, err := w.Write([]byte(manifest)); err != nil {
			logger.Error("Failed to write manifest", "error", err)
		}
	})

	// Start server
	server := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}
