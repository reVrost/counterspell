// Package ui contains the embedded dist directory.
package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distDir embed.FS

// DistFS returns the embedded dist directory without dist prefix.
var DistDirFs, _ = fs.Sub(distDir, "dist")

// Static returns the embedded static filesystem
func Static() (fs.FS, error) {
	return fs.Sub(distDir, "dist")
}
