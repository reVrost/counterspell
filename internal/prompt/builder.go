package prompt

import (
	"os"
	"strings"
)

// Builder assembles prompt text from lines and sections.
type Builder struct {
	parts []string
}

// NewBuilder creates a new prompt builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// AddLine appends a single line if it's not empty.
func (b *Builder) AddLine(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	b.parts = append(b.parts, line)
}

// AddSection appends a titled section if the body is not empty.
func (b *Builder) AddSection(title, body string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}
	if strings.TrimSpace(title) != "" {
		b.parts = append(b.parts, title)
	}
	b.parts = append(b.parts, body)
}

// AddFileSection reads a file and appends it as a section.
func (b *Builder) AddFileSection(title, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	b.AddSection(title, string(data))
	return nil
}

// String returns the assembled prompt.
func (b *Builder) String() string {
	return strings.Join(b.parts, "\n\n")
}
