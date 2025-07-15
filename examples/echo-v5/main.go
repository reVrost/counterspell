package main

import (
	"net/http"

	echov5 "github.com/labstack/echo/v5"
	"github.com/revrost/counterspell"
	"github.com/rs/zerolog/log"
)

func main() {
	// Example 3: Using Echo v5 router with DuckDB backend
	log.Info().Msg("Starting Echo v5 server with Counterspell...")
	e := echov5.New()

	// Add Counterspell to Echo v5 router - uses DuckDB for storage
	_, err := counterspell.AddToEchoV5(e,
		counterspell.WithAuthToken("my-echo-v5-token"),
		counterspell.WithDBPath("counterspell_echo_v5.db"),
		counterspell.WithServiceName("counterspell-echo-v5-example"),
		counterspell.WithServiceVersion("2.0.0"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add Counterspell to Echo v5")
	}

	// Add your application routes
	e.GET("/hello", func(c echov5.Context) error {
		log.Ctx(c.Request().Context()).Debug().Str("user", "demo").Msg("Hello from Echo v5 with Counterspell!")
		log.Ctx(c.Request().Context()).Error().Msg("Hello from Echo v5 with Counterspell!")
		return c.String(http.StatusOK, "Hello from Echo v5 with Counterspell!")
	})

	// Add a route that demonstrates tracing
	e.GET("/slow", func(c echov5.Context) error {
		log.Ctx(c.Request().Context()).Info().Msg("Starting slow operation...")

		// Simulate some work
		for i := 0; i < 3; i++ {
			log.Ctx(c.Request().Context()).Debug().Int("step", i+1).Msg("Processing step")
		}

		log.Ctx(c.Request().Context()).Info().Msg("Slow operation completed")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Slow operation completed",
			"status":  "success",
		})
	})

	log.Info().Msg("Echo v5 server listening on :8082")
	log.Info().Msg("Counterspell UI available at: http://localhost:8082/counterspell/health")
	log.Info().Msg("Counterspell API available at: http://localhost:8082/counterspell/api/logs?secret=my-echo-v5-token")
	log.Info().Msg("Test endpoints: /hello, /slow")
	if err := e.Start(":8082"); err != nil {
		log.Printf("Echo v5 server stopped: %v", err)
	}
}
