package codex

// ConfigValue is a JSON-like value allowed for Codex config overrides.
type ConfigValue any

// ConfigObject is a JSON-like object for Codex config overrides.
type ConfigObject map[string]ConfigValue

// Options configures a Codex client.
type Options struct {
	// CodexPathOverride overrides the path to the codex CLI binary.
	CodexPathOverride string
	BaseURL           string
	APIKey            string
	// Config provides additional --config overrides for the Codex CLI.
	Config ConfigObject
	// Env overrides environment variables passed to the Codex CLI process.
	// When set, only these variables are provided (plus required Codex vars).
	Env map[string]string
}
