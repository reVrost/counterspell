package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/your-github-username/microscope"
)

func main() {
	// Example 1: Using Echo router
	log.Println("Starting Echo server with MicroScope...")
	e := echo.New()

	// Add MicroScope to Echo router
	if err := microscope.AddToEcho(e,
		microscope.WithAuthToken("my-secret-token"),
		microscope.WithDBPath("microscope_echo.db"),
	); err != nil {
		log.Fatal("Failed to add MicroScope to Echo:", err)
	}

	// Add your application routes
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Echo with MicroScope!")
	})

	go func() {
		log.Println("Echo server listening on :8080")
		log.Println("MicroScope UI available at: http://localhost:8080/microscope/health")
		log.Println("MicroScope API available at: http://localhost:8080/microscope/api/logs?secret=my-secret-token")
		if err := e.Start(":8080"); err != nil {
			log.Printf("Echo server stopped: %v", err)
		}
	}()

	// Example 2: Using standard library router (Go 1.22+)
	log.Println("Starting stdlib server with MicroScope...")
	mux := http.NewServeMux()

	// Add MicroScope to stdlib router
	if _, err := microscope.AddToStdlib(mux,
		microscope.WithAuthToken("my-other-secret-token"),
		microscope.WithDBPath("microscope_stdlib.db"),
	); err != nil {
		log.Fatal("Failed to add MicroScope to stdlib router:", err)
	}

	// Add your application routes
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from stdlib router with MicroScope!"))
	})

	go func() {
		log.Println("Stdlib server listening on :8081")
		log.Println("MicroScope UI available at: http://localhost:8081/microscope/health")
		log.Println("MicroScope API available at: http://localhost:8081/microscope/api/logs?secret=my-other-secret-token")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Printf("Stdlib server stopped: %v", err)
		}
	}()

	// Keep the main goroutine alive
	select {}
}
