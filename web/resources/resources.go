package resources

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticEmbed embed.FS

// StaticFS returns an http.FileSystem for serving embedded static files.
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		panic("resources: missing static directory: " + err.Error())
	}
	return http.FS(sub)
}
