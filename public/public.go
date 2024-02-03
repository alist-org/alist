package public

import "embed"

//go:embed all:dist
var Public embed.FS
