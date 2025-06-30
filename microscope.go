package microscope

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/your-github-username/microscope/internal/microscope"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

// config holds the configuration for MicroScope
type config struct {
	dbPath    string
	authToken string
}

// Option represents a configuration option for MicroScope
type Option func(*config)

// WithDBPath sets the database path
func WithDBPath(path string) Option {
	return func(c *config) {
		c.dbPath = path
	}
}

// WithAuthToken sets the authentication token
func WithAuthToken(token string) Option {
	return func(c *config) {
		c.authToken = token
	}
}

// MicroScope holds the internal state for the observability system
type MicroScope struct {
	db            *sql.DB
	tracerProvider *trace.TracerProvider
	logWriter     *microscope.SQLiteLogWriter
}

// Install initializes MicroScope with the provided Echo instance
func Install(e *echo.Echo, opts ...Option) error {
	cfg := &config{
		dbPath:    "microscope.db",
		authToken: os.Getenv("MICROSCOPE_AUTH_TOKEN"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.authToken == "" {
		return fmt.Errorf("auth token is required: set MICROSCOPE_AUTH_TOKEN environment variable or use WithAuthToken option")
	}

	db, err := initDatabase(cfg.dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	ms, err := setupObservability(db)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to setup observability: %w", err)
	}

	configureGlobalLogger(ms.logWriter)
	registerRoutes(e, db, cfg.authToken)
	registerShutdownHook(e, ms)

	log.Info().Str("db_path", cfg.dbPath).Msg("MicroScope installed successfully")
	return nil
}

func initDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetBaseFS(migrationsFS)
	
	if err := goose.SetDialect("sqlite3"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func setupObservability(db *sql.DB) (*MicroScope, error) {
	exporter := microscope.NewSQLiteSpanExporter(db)
	logWriter := microscope.NewSQLiteLogWriter(db)

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("microscope-app"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return &MicroScope{
		db:            db,
		tracerProvider: tp,
		logWriter:     logWriter,
	}, nil
}

func configureGlobalLogger(logWriter *microscope.SQLiteLogWriter) {
	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		logWriter,
	)

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger().
		Hook(tracingHook{})
}

type tracingHook struct{}

func (h tracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if span := oteltrace.SpanFromContext(context.Background()); span.SpanContext().IsValid() {
		e.Str("trace_id", span.SpanContext().TraceID().String())
		e.Str("span_id", span.SpanContext().SpanID().String())
	}
}

func registerRoutes(e *echo.Echo, db *sql.DB, authToken string) {
	handler := microscope.NewAPIHandler(db)

	microscopeGroup := e.Group("/microscope")

	apiGroup := microscopeGroup.Group("/api", middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == authToken, nil
	}))

	apiGroup.GET("/logs", handler.QueryLogs)
	apiGroup.GET("/traces", handler.QueryTraces)
	apiGroup.GET("/traces/:trace_id", handler.GetTraceDetails)

	microscopeGroup.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "microscope",
		})
	})
}

func registerShutdownHook(e *echo.Echo, ms *MicroScope) {
	e.Server.RegisterOnShutdown(func() {
		log.Info().Msg("MicroScope shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ms.tracerProvider.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown tracer provider")
		}

		if err := ms.logWriter.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log writer")
		}

		if err := ms.db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database")
		}

		log.Info().Msg("MicroScope shutdown complete")
	})
} 