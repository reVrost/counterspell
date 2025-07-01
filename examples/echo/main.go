package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/your-github-username/microscope"
)

func main() {
	// Example 1: Using Echo router
	log.Info().Msg("Starting Echo server with Microscope...")
	e := echo.New()

	// Add Microscope to Echo router
	if err := microscope.AddToEcho(e,
		microscope.WithAuthToken("my-secret-token"),
		microscope.WithDBPath("microscope_echo.db"),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to add Microscope to Echo")
	}

	// Add your application routes
	e.GET("/hello", func(c echo.Context) error {
		log.Ctx(c.Request().Context()).Debug().Msg("Hello from Echo with Microscope!")
		log.Ctx(c.Request().Context()).Error().Msg("Hello from Echo with Microscope!")
		return c.String(http.StatusOK, "Hello from Echo with Microscope!")
	})

	log.Info().Msg("Echo server listening on :8080")
	log.Info().Msg("Microscope UI available at: http://localhost:8080/microscope/health")
	log.Info().Msg("Microscope API available at: http://localhost:8080/microscope/api/logs?secret=my-secret-token")
	if err := e.Start(":8080"); err != nil {
		log.Printf("Echo server stopped: %v", err)
	}
}
