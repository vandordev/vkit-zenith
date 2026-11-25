package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func Environment() map[string]string {
	out := make(map[string]string)
	for _, item := range os.Environ() {
		key, value, _ := strings.Cut(item, "=")
		out[key] = value
	}
	return out
}

// DevelopmentEnvironment explicitly reads only root/.env. It does not mutate
// process state, and already injected values (including empty ones) win.
func DevelopmentEnvironment(root string, injected map[string]string) (map[string]string, error) {
	file := filepath.Join(root, ".env")
	out, err := godotenv.Read(file)
	if err != nil {
		return nil, fmt.Errorf("Configuration %s [development] document: cannot parse root env file", file)
	}
	if injected == nil {
		injected = Environment()
	}
	for key, value := range injected {
		out[key] = value
	}
	return out, nil
}
