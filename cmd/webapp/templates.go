package main

import (
	"embed"

	"github.com/jaitaiwan/haproxy-manager/internal/util"
)

//go:embed static
var staticFiles embed.FS

type DefaultViewData struct {
	Title  string
	Flags  *util.FeatureFlags
	Active string
}
