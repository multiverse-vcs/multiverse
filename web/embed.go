package web

import (
	"embed"
)

//go:embed public/*
var Public embed.FS

//go:embed html/*
var HTML embed.FS
