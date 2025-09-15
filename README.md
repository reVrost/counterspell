# Counterspell

**⚠️ This project is a work in progress and is not yet ready for production use. ⚠️**

A lightweight, embedded observability tool for Go applications that provides OpenTelemetry tracing and logging capabilities with a local SQLite database backend and REST API for data querying.

## What it is

- Fast and easy to get started
- Gives you observability UI with the greatest of ease
- Embedded observability with otel, zerolog (uses sqlite)
- Means no external dependencies, no xtra docker containers
- Writes logs on a separate goroutine, so your app is not affected
- Will add features for LLM ops soon!

## Installation

```bash
go get github.com/revrost/counterspell
```

## Todo

- [ ] Agent configuration framework
- [ ] Openrouter integration
- [ ] Lightweight execution runtime via goroutine (inspired by cadence/go-workflow)
- [ ] Orchestrator-Executor MVP via cli
- [ ] Openapi streaming spec

## Quick Start

The simplest way to add Counterspell to your application:

```go
func main() {
	// Example 1: Using Echo router
	log.Info().Msg("Starting Echo server with Counterspell...")
	e := echo.New()

	// Use echo
	e.Use(otelecho.Middleware("counterspell-example"))

	// Add Counterspell to Echo router
	if err := counterspell.AddToEcho(e,
		counterspell.WithAuthToken("my-secret-token"),
		counterspell.WithDBPath("counterspell_echo.db"),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to add Counterspell to Echo")
	}

	// Add your application routes
	e.GET("/hello", func(c echo.Context) error {
		log.Ctx(c.Request().Context()).Debug().Str("user", "demo").Msg("Hello from Echo with Counterspell!")
		log.Ctx(c.Request().Context()).Error().Msg("Hello from Echo with Counterspell!")
		return c.String(http.StatusOK, "Hello from Echo with Counterspell!")
	})

	log.Info().Msg("Echo server listening on :8080")
	log.Info().Msg("Counterspell UI available at: http://localhost:8080/counterspell/health")
	log.Info().Msg("Counterspell API available at: http://localhost:8080/counterspell/api/logs?secret=my-secret-token")
	if err := e.Start(":8080"); err != nil {
		log.Printf("Echo server stopped: %v", err)
	}
}
```

## API Endpoints

All API endpoints require authentication via the `Authorization: Bearer <token>` header or `auth` query parameter.

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

- GitHub Issues: Report bugs and request features
- Documentation: See the [docs](./docs) directory for detailed guides
- Examples: Check the [examples](./examples) directory for more usage patterns
