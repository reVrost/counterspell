package tunnel

import "context"

// Provider defines the minimal contract required to expose
// a local HTTP service via a remote tunnel.
type Provider interface {

	// Name returns a stable provider identifier.
	// Example: "cloudflare", "ngrok"
	Name() string

	// Start launches the tunnel process.
	// This must be non-blocking.
	Start(ctx context.Context, cfg StartConfig) error

	// Stop gracefully terminates the tunnel.
	Stop(ctx context.Context) error

	// Status returns the current runtime state.
	Status(ctx context.Context) (RuntimeStatus, error)
}
