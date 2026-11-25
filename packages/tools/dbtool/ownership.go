package dbtool

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type ownershipKey struct{}

// WithOwnershipJournal opts an owned test into acquisition-time crash recovery.
// The launcher creates the private journal outside its disposable HOME. Runtime
// packages never import dbtool, and there is no ambient configuration read here.
func WithOwnershipJournal(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ownershipKey{}, path)
}

// RecordOwnership fsyncs safe identity only, before the caller can proceed.
// A failed journal write fails acquisition; never log DSNs, commands or secrets.
func RecordOwnership(ctx context.Context, kind, id string) error {
	path, _ := ctx.Value(ownershipKey{}).(string)
	if path == "" {
		return nil // existing non-Auth fixture consumers retain their own lifecycle.
	}
	if !strings.HasPrefix(path, "/tmp/opencode/auth-") || !strings.HasSuffix(path, "/ownership.jsonl") {
		return errors.New("owned resource: invalid journal")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return errors.New("owned resource: journal unavailable")
	}
	defer f.Close()
	data, err := json.Marshal(struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	}{kind, id})
	if err == nil {
		_, err = f.Write(append(data, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		return errors.New("owned resource: journal write failed")
	}
	return nil
}
