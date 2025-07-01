package microscope

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/your-github-username/microscope/internal/microscope"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

// config holds the configuration for Microscope
type config struct {
	dbPath        string
	authToken     string
	serviceName   string
	serviceVesion string
}

// Option represents a configuration option for Microscope
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

// WithServiceName sets the service name
func WithServiceName(name string) Option {
	return func(c *config) {
		c.serviceName = name
	}
}

// WithServiceVersion sets the service version
func WithServiceVersion(version string) Option {
	return func(c *config) {
		c.serviceVesion = version
	}
}

// Microscope holds the internal state for the observability system
type Microscope struct {
	db             *sql.DB
	tracerProvider *trace.TracerProvider
	logWriter      *microscope.SQLiteLogWriter
}

// Install is deprecated. Use AddToEcho instead.
// This function is kept for backward compatibility.
func Install(e *echo.Echo, opts ...Option) error {
	return AddToEcho(e, opts...)
}

// AddToEcho initializes Microscope with the provided Echo instance
func AddToEcho(e *echo.Echo, opts ...Option) error {
	cfg := &config{
		dbPath:        "microscope.db",
		authToken:     os.Getenv("MICROSCOPE_AUTH_TOKEN"),
		serviceName:   "microscope-app",
		serviceVesion: "1.0.0",
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

	ms, err := setupObservability(db, cfg.serviceName, cfg.serviceVesion)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to setup observability: %w", err)
	}

	configureGlobalLogger(ms.logWriter)
	e.Use(otelecho.Middleware(cfg.serviceName))
	e.Use(loggerMiddleware)
	registerEchoRoutes(e, db, cfg.authToken)
	registerEchoShutdownHook(e, ms)

	log.Info().Str("db_path", cfg.dbPath).Msg("Microscope installed successfully")
	return nil
}

// AddToStdlib initializes Microscope with the provided standard library ServeMux
func AddToStdlib(mux *http.ServeMux, opts ...Option) (*Microscope, error) {
	cfg := &config{
		dbPath:        "microscope.db",
		authToken:     os.Getenv("MICROSCOPE_AUTH_TOKEN"),
		serviceName:   "microscope-app",
		serviceVesion: "1.0.0",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.authToken == "" {
		return nil, fmt.Errorf("auth token is required: set MICROSCOPE_AUTH_TOKEN environment variable or use WithAuthToken option")
	}

	db, err := initDatabase(cfg.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	ms, err := setupObservability(db, cfg.serviceName, cfg.serviceVesion)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup observability: %w", err)
	}

	configureGlobalLogger(ms.logWriter)
	registerStdlibRoutes(mux, db, cfg.authToken)

	log.Info().Str("db_path", cfg.dbPath).Msg("Microscope installed successfully")
	return ms, nil
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

func setupObservability(db *sql.DB, serviceName, serviceVersion string) (*Microscope, error) {
	exporter := microscope.NewSQLiteSpanExporter(db)
	logWriter := microscope.NewSQLiteLogWriter(db)

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
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

	return &Microscope{
		db:             db,
		tracerProvider: tp,
		logWriter:      logWriter,
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
		Logger()
}

type tracingHook struct {
	ctx context.Context
}

func (h tracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if span := oteltrace.SpanFromContext(h.ctx); span.SpanContext().IsValid() {
		e.Str("trace_id", span.SpanContext().TraceID().String())
		e.Str("span_id", span.SpanContext().SpanID().String())
	}
}

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create a new logger with a tracing hook that has the request context.
		loggerWithTrace := log.Logger.Hook(tracingHook{ctx: c.Request().Context()})

		// Create a new context with the new logger.
		ctxWithLogger := loggerWithTrace.WithContext(c.Request().Context())

		// Create a new request with the new context.
		req := c.Request().WithContext(ctxWithLogger)

		// Set the new request in the Echo context.
		c.SetRequest(req)

		return next(c)
	}
}

func registerEchoRoutes(e *echo.Echo, db *sql.DB, authToken string) {
	handler := microscope.NewAPIHandler(db)

	microscopeGroup := e.Group("/microscope")

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

	apiGroup := microscopeGroup.Group("/api", secretAuth)

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

func registerStdlibRoutes(mux *http.ServeMux, db *sql.DB, authToken string) {
	apiHandler := microscope.NewAPIHandler(db)

	// Create auth middleware wrapper for stdlib
	withAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			secret := r.URL.Query().Get("secret")
			if secret == "" {
				http.Error(w, "secret query parameter is required", http.StatusBadRequest)
				return
			}
			if secret != authToken {
				http.Error(w, "invalid secret", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}

	// Create native stdlib handlers instead of trying to adapt Echo handlers
	logsHandler := withAuth(func(w http.ResponseWriter, r *http.Request) {
		// Call the API logic through a temporary Echo context wrapper
		e := echo.New()
		c := e.NewContext(r, httptest.NewRecorder())
		c.SetRequest(r)

		// Set query params on the context
		for key, values := range r.URL.Query() {
			for _, value := range values {
				c.QueryParams()[key] = append(c.QueryParams()[key], value)
			}
		}

		// Call the handler
		if err := apiHandler.QueryLogs(c); err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				http.Error(w, httpErr.Message.(string), httpErr.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Copy response from echo recorder to actual response
		recorder := c.Response().Writer.(*httptest.ResponseRecorder)
		for key, values := range recorder.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})

	tracesHandler := withAuth(func(w http.ResponseWriter, r *http.Request) {
		e := echo.New()
		c := e.NewContext(r, httptest.NewRecorder())
		c.SetRequest(r)

		// Set query params on the context
		for key, values := range r.URL.Query() {
			for _, value := range values {
				c.QueryParams()[key] = append(c.QueryParams()[key], value)
			}
		}

		if err := apiHandler.QueryTraces(c); err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				http.Error(w, httpErr.Message.(string), httpErr.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		recorder := c.Response().Writer.(*httptest.ResponseRecorder)
		for key, values := range recorder.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})

	traceDetailsHandler := withAuth(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace_id from URL path
		path := r.URL.Path
		traceID := ""
		if len(path) > len("/microscope/api/traces/") {
			traceID = path[len("/microscope/api/traces/"):]
		}

		if traceID == "" {
			http.Error(w, "trace_id is required", http.StatusBadRequest)
			return
		}

		e := echo.New()
		c := e.NewContext(r, httptest.NewRecorder())
		c.SetRequest(r)
		c.SetParamNames("trace_id")
		c.SetParamValues(traceID)

		if err := apiHandler.GetTraceDetails(c); err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				http.Error(w, httpErr.Message.(string), httpErr.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		recorder := c.Response().Writer.(*httptest.ResponseRecorder)
		for key, values := range recorder.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "microscope",
		})
	}

	// Register routes
	mux.HandleFunc("/microscope/api/logs", logsHandler)
	mux.HandleFunc("/microscope/api/traces", tracesHandler)
	mux.HandleFunc("/microscope/api/traces/", traceDetailsHandler) // Note: trailing slash for path pattern matching
	mux.HandleFunc("/microscope/health", healthHandler)
}

func registerEchoShutdownHook(e *echo.Echo, ms *Microscope) {
	e.Server.RegisterOnShutdown(func() {
		log.Info().Msg("Microscope shutting down...")

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

		log.Info().Msg("Microscope shutdown complete")
	})
}
