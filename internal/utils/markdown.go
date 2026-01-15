package utils

import (
	"bytes"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// RenderMarkdownHTML converts markdown to HTML with headings rendered as bold (no size change).
func RenderMarkdownHTML(content string) string {
	md := goldmark.New(
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(&headingAsBoldRenderer{}, 100),
			),
			goldhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return html.EscapeString(content)
	}
	return buf.String()
}

// headingAsBoldRenderer renders headings as bold text instead of h1-h6 tags.
type headingAsBoldRenderer struct{}

func (r *headingAsBoldRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *headingAsBoldRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<p><strong>")
	} else {
		_, _ = w.WriteString("</strong></p>\n")
	}
	return ast.WalkContinue, nil
}

// ExtractTitleFromMarkdown extracts the first heading from markdown content.
// It looks for # heading (H1) first, then ## (H2), etc.
// If no heading is found, it returns the first line or a default.
func ExtractTitleFromMarkdown(content string) string {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Find the heading level and extract text
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// No heading found, return first non-empty line as title
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Limit title to 60 characters
			if len(trimmed) > 60 {
				return trimmed[:57] + "..."
			}
			return trimmed
		}
	}

	return "Untitled Task"
}
