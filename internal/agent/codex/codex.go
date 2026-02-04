package codex

// Codex is the main client for interacting with the Codex agent.
type Codex struct {
	exec    *Exec
	options Options
}

func New(options Options) *Codex {
	return &Codex{
		exec:    NewExec(options.CodexPathOverride, options.Env, options.Config),
		options: options,
	}
}

func (c *Codex) StartThread(options ThreadOptions) *Thread {
	return NewThread(c.exec, c.options, options, "")
}

func (c *Codex) ResumeThread(id string, options ThreadOptions) *Thread {
	return NewThread(c.exec, c.options, options, id)
}
