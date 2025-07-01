package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/your-github-username/counterspell"
)

func main() {
	// Example 2: Using standard library router (Go 1.22+)
	log.Info().Msg("Starting stdlib server with Counterspell...")
	mux := http.NewServeMux()

	// Add Counterspell to stdlib router
	if _, err := counterspell.AddToStdlib(mux,
		counterspell.WithAuthToken("my-other-secret-token"),
		counterspell.WithDBPath("counterspell_stdlib.db"),
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to add Counterspell to stdlibbj")
	}

	// Add your application routes
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Ctx(r.Context()).Debug().Str("user", "demo").Msg("Hello from stdlib router with Counterspell!")
		log.Ctx(r.Context()).Error().Msg("Hello from stdlib router with Counterspell!")
		w.Write([]byte("Hello from stdlib router with Counterspell!"))
	})

	log.Info().Msg("Stdlib server listening on :8081")
	log.Info().Msg("Counterspell UI available at: http://localhost:8081/counterspell/health")
	log.Info().Msg("Counterspell API available at: http://localhost:8081/counterspell/api/logs?secret=my-other-secret-token")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Printf("Stdlib server stopped: %v", err)
	}
}
