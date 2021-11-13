package public

import "embed"

//go:embed *
var Public embed.FS

////go:embed index.html
//var Index embed.FS
//
////go:embed assets/**
//var Assets embed.FS
