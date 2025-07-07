package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distDir embed.FS

// DistDirFS is the embedded dist directory without dist prefix
var DistDirFS, _ = fs.Sub(distDir, "dist")
