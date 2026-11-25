package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ApplicationRoot accepts trusted bootstrap input, never HTTP caller input.
// Deployment supplies /app explicitly; development commands supply repo root.
func ApplicationRoot(root string) (string, error) {
	if root == "" {
		root = os.Getenv("ZENITH_APPLICATION_ROOT")
	}
	if root == "" {
		return "", fmt.Errorf("Trusted application root required for configuration")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("Invalid trusted application root")
	}
	return abs, nil
}
