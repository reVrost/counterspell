package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var FS embed.FS

// Static returns the embedded static filesystem
func Static() (fs.FS, error) {
	return fs.Sub(FS, "dist")
}
