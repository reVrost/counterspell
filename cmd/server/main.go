package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/revrost/counterspell"
	"github.com/revrost/counterspell/pkg/gen/proto/counterspell/v1/counterspellv1connect"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func ConnectHandler(path string, handler http.Handler) (string, echo.HandlerFunc) {
	path = path + "*"
	return path, echo.WrapHandler(handler)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	// Initialize database
	db, err := initDB("counterspell.db")
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize observability
	cleanup, err := initObservability(db)
	if err != nil {
		slog.Error("Failed to initialize observability", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add tracing middleware
	e.Use(tracingMiddleware())

	// Initialize API handlers
	apiHandler := counterspell.NewAPIHandler(db)

	// Get auth token from environment
	authToken := os.Getenv("COUNTERSPELL_AUTH_TOKEN")
	if authToken == "" {
		authToken = "dev-token" // Default for development
	}

	// Custom middleware to check for secret query parameter
	secretAuth := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			secret := c.QueryParam("secret")
			if secret == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "secret query parameter is required")
			}
			if secret != authToken {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid secret")
			}
			return next(c)
		}
	}

	// Register counterspell API routes with authentication
	api := e.Group("/counterspell/api", secretAuth)
	api.GET("/logs", apiHandler.QueryLogs)
	api.GET("/traces", apiHandler.QueryTraces)
	api.GET("/traces/:trace_id", apiHandler.GetTraceDetails)
	api.POST("/chat", func(c echo.Context) error {
		// This is a placeholder implementation.
		// In a real application, you would call your LLM here.
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		// Echo back the last user message for now
		var lastUserMessage string
		for i := len(body.Messages) - 1; i >= 0; i-- {
			if body.Messages[i].Role == "user" {
				lastUserMessage = body.Messages[i].Content
				break
			}
		}

		return c.String(http.StatusOK, "Echo: "+lastUserMessage)
	})

	// Add health endpoint (no auth required)
	e.GET("/counterspell/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "counterspell",
		})
	})

	// Add example route for testing
	e.GET("/hello", helloHandler)
	e.GET("/slow", slowHandler)
	e.GET("/error", errorHandler)

	service := counterspell.NewService(db)
	// Connect RPC Handlers
	path, handlers := ConnectHandler(
		counterspellv1connect.NewServiceHandler(
			service,
			// connect.WithInterceptors(market.AuthInterceptor(server.app)),
		))
	e.Any(
		path, handlers,
		// UserActivityLogger(server.app),
	)

	// Start server
	slog.Info("Server starting on :8989")
	slog.Info("Counterspell API available at /counterspell/api")
	slog.Info("Example routes: /hello, /slow, /error")

	// Graceful shutdown
	go func() {
		if err := e.Start(":8989"); err != nil {
			slog.Error("Server stopped", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited")
}

// initDB initializes the SQLite database and runs migrations
func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database initialized successfully")
	return db, nil
}

// initObservability sets up OpenTelemetry tracing and logging
func initObservability(db *sql.DB) (func(), error) {
	// Create SQLite exporter
	exporter := counterspell.NewSQLiteSpanExporter(db)

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("counterspell-example"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Initialize logging
	logWriter := counterspell.NewSQLiteLogWriter(db)

	// Configure slog with multiple outputs
	multiHandler := slogmulti.Fanout(
		slog.NewTextHandler(os.Stdout, nil), // Console output for development
		slog.NewJSONHandler(logWriter, nil), // SQLite storage
	)

	// Configure global logger with trace context handler
	logger := slog.New(multiHandler).
		With(slog.String("timestamp", time.Now().Format(time.RFC3339)))

	// Set as default logger
	slog.SetDefault(logger)

	slog.Info("Observability initialized successfully")

	// Return cleanup function
	return func() {
		slog.Info("Shutting down observability...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown tracer provider", "error", err)
		}

		if err := logWriter.Close(); err != nil {
			slog.Error("Failed to close log writer", "error", err)
		}

		slog.Info("Observability shutdown complete")
	}, nil
}

// tracingMiddleware adds OpenTelemetry tracing to HTTP requests
func tracingMiddleware() echo.MiddlewareFunc {
	tracer := otel.Tracer("counterspell-http")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Start a new span for the HTTP request
			ctx, span := tracer.Start(req.Context(),
				fmt.Sprintf("%s %s", req.Method, req.URL.Path),
				oteltrace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.scheme", req.URL.Scheme),
					attribute.String("http.host", req.Host),
				),
			)
			defer span.End()

			// Update request context
			c.SetRequest(req.WithContext(ctx))

			// Call next handler
			err := next(c)

			// Set span status based on response
			if err != nil {
				span.RecordError(err)
				if httpErr, ok := err.(*echo.HTTPError); ok {
					span.SetAttributes(attribute.Int("http.status_code", httpErr.Code))
				}
			} else {
				span.SetAttributes(attribute.Int("http.status_code", c.Response().Status))
			}

			return err
		}
	}
}

// Example handlers for testing

func helloHandler(c echo.Context) error {
	tracer := otel.Tracer("counterspell-example")
	ctx, span := tracer.Start(c.Request().Context(), "hello-processing")
	defer span.End()

	// Update the request context for logging
	c.SetRequest(c.Request().WithContext(ctx))

	slog.Info("Processing hello request", "user", "demo-user", "endpoint", "/hello")

	span.SetAttributes(
		attribute.String("user", "demo-user"),
		attribute.String("greeting", "hello"),
	)

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	slog.Info("Hello request completed successfully")

	return c.JSON(200, map[string]string{
		"message":  "Hello, World!",
		"trace_id": span.SpanContext().TraceID().String(),
	})
}

func slowHandler(c echo.Context) error {
	tracer := otel.Tracer("counterspell-example")
	ctx, span := tracer.Start(c.Request().Context(), "slow-processing")
	defer span.End()

	c.SetRequest(c.Request().WithContext(ctx))

	slog.Info("Starting slow operation")

	// Simulate slow work with sub-spans
	for i := range 3 {
		_, subSpan := tracer.Start(ctx, fmt.Sprintf("slow-step-%d", i+1))

		slog.Info("Processing slow step", "step", i+1)

		time.Sleep(200 * time.Millisecond)
		subSpan.End()
	}

	slog.Info("Slow operation completed")

	return c.JSON(200, map[string]any{
		"message":     "Slow operation completed",
		"duration_ms": 600,
		"trace_id":    span.SpanContext().TraceID().String(),
	})
}

func errorHandler(c echo.Context) error {
	tracer := otel.Tracer("counterspell-example")
	ctx, span := tracer.Start(c.Request().Context(), "error-simulation")
	defer span.End()

	c.SetRequest(c.Request().WithContext(ctx))

	slog.Error("Simulating an error condition", "error_type", "simulated_error")

	span.SetAttributes(
		attribute.String("error.type", "simulated_error"),
		attribute.Bool("error.expected", true),
	)

	// Record error on span
	err := fmt.Errorf("this is a simulated error for testing")
	span.RecordError(err)

	return echo.NewHTTPError(500, "Simulated error for testing Counterspell")
}
