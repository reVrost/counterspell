package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/your-github-username/microscope"
)

func main() {
	// Example 2: Using standard library router (Go 1.22+)
	log.Info().Msg("Starting stdlib server with Microscope...")
	mux := http.NewServeMux()

	// Add Microscope to stdlib router
	if _, err := microscope.AddToStdlib(mux,
		microscope.WithAuthToken("my-other-secret-token"),
		microscope.WithDBPath("microscope_stdlib.db"),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to add Microscope to stdlibbj")
	}

	// Add your application routes
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Ctx(r.Context()).Debug().Str("user", "demo").Msg("Hello from stdlib router with Microscope!")
		log.Ctx(r.Context()).Error().Msg("Hello from stdlib router with Microscope!")
		w.Write([]byte("Hello from stdlib router with Microscope!"))
	})

	log.Info().Msg("Stdlib server listening on :8081")
	log.Info().Msg("Microscope UI available at: http://localhost:8081/microscope/health")
	log.Info().Msg("Microscope API available at: http://localhost:8081/microscope/api/logs?secret=my-other-secret-token")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Printf("Stdlib server stopped: %v", err)
	}
}
