//go:build integration

package server

// Integration fixtures compose independent modules in one test process, so
// their service-only CLI flags are not registered against the shared flag set.
var root = new(string)
var development = new(bool)
