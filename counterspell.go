package counterspell

import (
	"context"
	"database/sql"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	echov4 "github.com/labstack/echo/v4"
	middlewarev4 "github.com/labstack/echo/v4/middleware"
	echov5 "github.com/labstack/echo/v5"
	middlewarev5 "github.com/labstack/echo/v5/middleware"
	_ "github.com/marcboeker/go-duckdb/v2"

	"github.com/revrost/counterspell/internal/counterspell"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// config holds the configuration for Counterspell
type config struct {
	dbPath         string
	authToken      string
	serviceName    string
	serviceVersion string
}

// Option represents a configuration option for Counterspell
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
		c.serviceVersion = version
	}
}

// Counterspell holds the internal state for the observability system
type Counterspell struct {
	db             *sql.DB
	tracerProvider *trace.TracerProvider
	logWriter      *counterspell.DuckDBLogWriter
}

// Install is deprecated. Use AddToEcho instead.
// This function is kept for backward compatibility.
func Install(e *echov4.Echo, opts ...Option) (*Counterspell, error) {
	return AddToEcho(e, opts...)
}

// AddToEcho initializes Counterspell with the provided Echo v4 instance
func AddToEcho(e *echov4.Echo, opts ...Option) (*Counterspell, error) {
	cfg := buildConfig(opts...)

	if cfg.authToken == "" {
		return nil, fmt.Errorf("auth token is required: set COUNTERSPELL_AUTH_TOKEN environment variable or use WithAuthToken option")
	}

	db, err := initDatabase(cfg.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	cs, err := setupObservability(db, cfg.serviceName, cfg.serviceVersion)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup observability: %w", err)
	}

	setupGlobalLogger(cs.logWriter)
	setupEchoV4Middleware(e)
	registerEchoV4Routes(e, db, cfg.authToken)
	registerEchoV4ShutdownHook(e, cs)

	log.Info().Str("db_path", cfg.dbPath).Msg("Counterspell installed successfully with Echo v4")
	return cs, nil
}

// AddToEchoV5 initializes Counterspell with the provided Echo v5 instance
func AddToEchoV5(e *echov5.Echo, opts ...Option) (*Counterspell, error) {
	cfg := buildConfig(opts...)

	if cfg.authToken == "" {
		return nil, fmt.Errorf("auth token is required: set COUNTERSPELL_AUTH_TOKEN environment variable or use WithAuthToken option")
	}

	db, err := initDatabase(cfg.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	cs, err := setupObservability(db, cfg.serviceName, cfg.serviceVersion)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup observability: %w", err)
	}

	setupGlobalLogger(cs.logWriter)
	setupEchoV5Middleware(e)
	registerEchoV5Routes(e, db, cfg.authToken)
	registerEchoV5ShutdownHook(e, cs)

	log.Info().Str("db_path", cfg.dbPath).Msg("Counterspell installed successfully with Echo v5")
	return cs, nil
}

// AddToStdlib initializes Counterspell with the provided standard library ServeMux
func AddToStdlib(mux *http.ServeMux, opts ...Option) (*Counterspell, error) {
	cfg := buildConfig(opts...)

	if cfg.authToken == "" {
		return nil, fmt.Errorf("auth token is required: set COUNTERSPELL_AUTH_TOKEN environment variable or use WithAuthToken option")
	}

	db, err := initDatabase(cfg.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	cs, err := setupObservability(db, cfg.serviceName, cfg.serviceVersion)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup observability: %w", err)
	}

	setupGlobalLogger(cs.logWriter)
	registerStdlibRoutes(mux, db, cfg.authToken)

	log.Info().Str("db_path", cfg.dbPath).Msg("Counterspell installed successfully")
	return cs, nil
}

// buildConfig creates a config with defaults and applies options
func buildConfig(opts ...Option) *config {
	cfg := &config{
		dbPath:         "counterspell.db",
		authToken:      os.Getenv("COUNTERSPELL_AUTH_TOKEN"),
		serviceName:    "counterspell-app",
		serviceVersion: "1.0.0",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func initDatabase(dbPath string) (*sql.DB, error) {
	// Use the db package's Open function which includes schema creation
	return db.Open(dbPath)
}

func setupObservability(db *sql.DB, serviceName, serviceVersion string) (*Counterspell, error) {
	exporter := counterspell.NewDuckDBSpanExporter(db)
	logWriter := counterspell.NewDuckDBLogWriter(db)

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

	return &Counterspell{
		db:             db,
		tracerProvider: tp,
		logWriter:      logWriter,
	}, nil
}

func setupGlobalLogger(logWriter *counterspell.DuckDBLogWriter) {
	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
		logWriter,
	)

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger()
}

func setupEchoV4Middleware(e *echov4.Echo) {
	// Add CORS middleware
	e.Use(middlewarev4.CORS())

	// Add request ID middleware
	e.Use(middlewarev4.RequestID())

	// Add recover middleware
	e.Use(middlewarev4.Recover())

	// Add our custom logger middleware
	e.Use(createLoggerMiddleware())
}

func setupEchoV5Middleware(e *echov5.Echo) {
	// Add CORS middleware
	e.Use(middlewarev5.CORS())

	// Add request ID middleware
	e.Use(middlewarev5.RequestID())

	// Add recover middleware
	e.Use(middlewarev5.Recover())

	// Add our custom logger middleware
	e.Use(createLoggerMiddlewareV5())
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

func createLoggerMiddleware() echov4.MiddlewareFunc {
	return func(next echov4.HandlerFunc) echov4.HandlerFunc {
		return func(c echov4.Context) error {
			// Create a new logger with a tracing hook that has the request context.
			loggerWithTrace := log.Logger.Hook(tracingHook{ctx: c.Request().Context()})

			// Replace the global logger with our context-aware logger for this request.
			c.Set("logger", &loggerWithTrace)

			return next(c)
		}
	}
}

func createLoggerMiddlewareV5() echov5.MiddlewareFunc {
	return func(next echov5.HandlerFunc) echov5.HandlerFunc {
		return func(c echov5.Context) error {
			// Create a new logger with a tracing hook that has the request context.
			loggerWithTrace := log.Logger.Hook(tracingHook{ctx: c.Request().Context()})

			// Replace the global logger with our context-aware logger for this request.
			c.Set("logger", &loggerWithTrace)

			return next(c)
		}
	}
}

func createAuthMiddleware(authToken string) echov4.MiddlewareFunc {
	return func(next echov4.HandlerFunc) echov4.HandlerFunc {
		return func(c echov4.Context) error {
			secret := c.QueryParam("secret")
			if secret == "" {
				return echov4.NewHTTPError(http.StatusBadRequest, "secret query parameter is required")
			}
			if secret != authToken {
				return echov4.NewHTTPError(http.StatusUnauthorized, "invalid secret")
			}
			return next(c)
		}
	}
}

func createAuthMiddlewareV5(authToken string) echov5.MiddlewareFunc {
	return func(next echov5.HandlerFunc) echov5.HandlerFunc {
		return func(c echov5.Context) error {
			secret := c.QueryParam("secret")
			if secret == "" {
				return echov5.NewHTTPError(http.StatusBadRequest, "secret query parameter is required")
			}
			if secret != authToken {
				return echov5.NewHTTPError(http.StatusUnauthorized, "invalid secret")
			}
			return next(c)
		}
	}
}

func registerEchoV4Routes(e *echov4.Echo, db *sql.DB, authToken string) {
	handler := counterspell.NewAPIHandler(db)
	authMiddleware := createAuthMiddleware(authToken)

	counterspellGroup := e.Group("/counterspell")

	// Serve UI static files
	counterspellGroup.Static("/", "ui/dist")

	// Create API group with authentication
	apiGroup := counterspellGroup.Group("/api", authMiddleware)
	apiGroup.GET("/logs", handler.QueryLogs)
	apiGroup.GET("/traces", handler.QueryTraces)
	apiGroup.GET("/traces/:trace_id", handler.GetTraceDetails)

	// Health endpoint (no auth required)
	counterspellGroup.GET("/health", createHealthHandler())
}

func createHealthHandler() echov4.HandlerFunc {
	return func(c echov4.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "counterspell",
		})
	}
}

func registerEchoV5Routes(e *echov5.Echo, db *sql.DB, authToken string) {
	handler := counterspell.NewAPIHandler(db)
	authMiddleware := createAuthMiddlewareV5(authToken)

	counterspellGroup := e.Group("/counterspell")

	// Serve UI static files
	counterspellGroup.StaticFS("/", ui.DistDirFS)

	// Create API group with authentication
	apiGroup := counterspellGroup.Group("/api", authMiddleware)
	apiGroup.GET("/logs", createEchoV5HandlerWrapper(handler.QueryLogs))
	apiGroup.GET("/traces", createEchoV5HandlerWrapper(handler.QueryTraces))
	apiGroup.GET("/traces/:trace_id", createEchoV5HandlerWrapper(handler.GetTraceDetails))

	// Health endpoint (no auth required)
	counterspellGroup.GET("/health", createHealthHandlerV5())
}

// createEchoV5HandlerWrapper creates a wrapper to use Echo v4 handlers with Echo v5
func createEchoV5HandlerWrapper(v4Handler echov4.HandlerFunc) echov5.HandlerFunc {
	return func(c echov5.Context) error {
		// Create a new Echo v4 instance and context
		e4 := echov4.New()
		c4 := e4.NewContext(c.Request(), &echoV5ResponseWriter{c: c})

		// Copy path parameters from v5 to v4
		// In Echo v5, we use PathParam to get individual parameters
		if traceID := c.PathParam("trace_id"); traceID != "" {
			c4.SetParamNames("trace_id")
			c4.SetParamValues(traceID)
		}

		// Call the v4 handler
		return v4Handler(c4)
	}
}

// echoV5ResponseWriter adapts Echo v5 response to Echo v4 format
type echoV5ResponseWriter struct {
	c echov5.Context
}

func (w *echoV5ResponseWriter) Header() http.Header {
	return w.c.Response().Header()
}

func (w *echoV5ResponseWriter) Write(b []byte) (int, error) {
	return w.c.Response().Write(b)
}

func (w *echoV5ResponseWriter) WriteHeader(statusCode int) {
	w.c.Response().WriteHeader(statusCode)
}

func createHealthHandlerV5() echov5.HandlerFunc {
	return func(c echov5.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "counterspell",
		})
	}
}

func registerStdlibRoutes(mux *http.ServeMux, db *sql.DB, authToken string) {
	apiHandler := counterspell.NewAPIHandler(db)
	withAuth := createStdlibAuthWrapper(authToken)

	// Register API routes
	mux.HandleFunc("/counterspell/api/logs", withAuth(createStdlibLogsHandler(apiHandler)))
	mux.HandleFunc("/counterspell/api/traces", withAuth(createStdlibTracesHandler(apiHandler)))
	mux.HandleFunc("/counterspell/api/traces/", withAuth(createStdlibTraceDetailsHandler(apiHandler)))
	mux.HandleFunc("/counterspell/health", createStdlibHealthHandler())
}

func createStdlibAuthWrapper(authToken string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
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
}

func createStdlibLogsHandler(apiHandler *counterspell.APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleStdlibEchoV4Adapter(w, r, apiHandler.QueryLogs)
	}
}

func createStdlibTracesHandler(apiHandler *counterspell.APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleStdlibEchoV4Adapter(w, r, apiHandler.QueryTraces)
	}
}

func createStdlibTraceDetailsHandler(apiHandler *counterspell.APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trace_id from URL path
		path := r.URL.Path
		parts := strings.Split(path, "/")
		var traceID string
		for i, part := range parts {
			if part == "traces" && i+1 < len(parts) {
				traceID = parts[i+1]
				break
			}
		}

		// Create Echo v4 context with trace_id parameter
		e := echov4.New()
		c := e.NewContext(r, httptest.NewRecorder())
		c.SetRequest(r)
		c.SetParamNames("trace_id")
		c.SetParamValues(traceID)

		// Call the handler
		if err := apiHandler.GetTraceDetails(c); err != nil {
			handleEchoError(w, err)
			return
		}

		// Copy response
		copyEchoV4Response(w, c)
	}
}

func createStdlibHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "counterspell",
		})
	}
}

func handleStdlibEchoV4Adapter(w http.ResponseWriter, r *http.Request, handler echov4.HandlerFunc) {
	e := echov4.New()
	c := e.NewContext(r, httptest.NewRecorder())
	c.SetRequest(r)

	if err := handler(c); err != nil {
		handleEchoError(w, err)
		return
	}

	copyEchoV4Response(w, c)
}

func handleEchoError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(*echov4.HTTPError); ok {
		http.Error(w, httpErr.Message.(string), httpErr.Code)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func copyEchoV4Response(w http.ResponseWriter, c echov4.Context) {
	recorder := c.Response().Writer.(*httptest.ResponseRecorder)
	for key, values := range recorder.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(recorder.Code)
	w.Write(recorder.Body.Bytes())
}

func registerEchoV4ShutdownHook(e *echov4.Echo, cs *Counterspell) {
	// Echo v4 has a Server field for shutdown hooks
	if e.Server != nil {
		e.Server.RegisterOnShutdown(func() {
			shutdownCounterspell(cs)
		})
	}
}

func registerEchoV5ShutdownHook(e *echov5.Echo, cs *Counterspell) {
	// Echo v5 might handle shutdown differently
	// For now, we'll use a simple approach by registering a cleanup function
	// This is a simplified approach - in real applications, you might want to
	// handle shutdown through the HTTP server directly
	go func() {
		// This is a placeholder - in real Echo v5, there might be a different shutdown API
		// For now, we'll let the application handle cleanup through other means
		_ = cs // Keep reference to prevent cleanup
	}()
}

func shutdownCounterspell(cs *Counterspell) {
	log.Info().Msg("Counterspell shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cs.tracerProvider.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown tracer provider")
	}

	if err := cs.logWriter.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close log writer")
	}

	if err := cs.db.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database")
	}

	log.Info().Msg("Counterspell shutdown complete")
}
