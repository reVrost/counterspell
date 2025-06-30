package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/your-github-username/microscope"
	"go.opentelemetry.io/otel"
)

func main() {
	e := echo.New()

	// --- MICROSCOPE: ONE-LINER INSTALLATION ---
	// The auth token is set via environment variable:
	// $ export MICROSCOPE_AUTH_TOKEN="my-secret-token"
	if err := microscope.Install(e); err != nil {
		log.Fatal().Err(err).Msg("Failed to install MicroScope")
	}
	// That's it. It's installed, configured, and will shut down gracefully.

	// For demonstration, you can also configure with options:
	/*
	if err := microscope.Install(
		e,
		microscope.WithAuthToken("my-super-secret-token"),
		microscope.WithDBPath("./data/observability.db"),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to install MicroScope")
	}
	*/
	// --- END OF INTEGRATION ---

	// The application's tracer and logger are now auto-configured.
	tracer := otel.Tracer("my-app")

	e.GET("/hello", func(c echo.Context) error {
		_, span := tracer.Start(c.Request().Context(), "hello-handler")
		defer span.End()

		// This log automatically has trace_id and will be written to SQLite.
		log.Info().Str("user", "frodo").Msg("A request has been received")
		return c.String(200, "Hello, World!")
	})

	log.Info().Msg("Server starting on :1323")
	log.Info().Msg("MicroScope API available at /microscope/api")
	e.Logger.Fatal(e.Start(":1323"))
} 