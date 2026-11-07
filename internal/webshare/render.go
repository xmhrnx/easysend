package webshare

import (
	"embed"
	"html/template"
)

//go:embed pages/*.html
var pageFS embed.FS

var pages = template.Must(template.ParseFS(pageFS, "pages/*.html"))
