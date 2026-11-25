//go:build !integration

package server

import "flag"

var root = flag.String("root", "", "Trusted application root containing config/config.yaml")
var development = flag.Bool("development", false, "Explicitly load root .env; injected environment wins")
