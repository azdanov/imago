package templates

import "embed"

//go:embed *.tmpl.html layouts/*.tmpl.html
var FS embed.FS
