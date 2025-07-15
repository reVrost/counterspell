package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/revrost/counterspell"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func main() {
	// Example 1: Using Echo v4 router with DuckDB backend
	log.Info().Msg("Starting Echo server with Counterspell...")
	e := echo.New()

	// Use echo
	e.Use(otelecho.Middleware("counterspell-example"))

	// Add Counterspell to Echo router - uses DuckDB for storage
	_, err := counterspell.AddToEcho(e,
		counterspell.WithAuthToken("my-secret-token"),
		counterspell.WithDBPath("counterspell_echo.db"),
	)
	if err != nil {
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
