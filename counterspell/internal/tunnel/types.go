package tunnel

type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusError   Status = "error"
)

// StartConfig is provider-agnostic.
// The provider decides how to interpret AuthToken.
type StartConfig struct {
	// Public hostname, e.g. alice.counterspell.app
	PublicHostname string

	// Local service address, e.g. http://localhost:5713
	LocalAddr string

	// Opaque provider token (Cloudflare, ngrok, etc.)
	AuthToken string
}

type RuntimeStatus struct {
	Status Status
	PID    int
	Error  error
}
