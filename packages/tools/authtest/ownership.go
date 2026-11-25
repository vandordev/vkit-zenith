package authtest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/vandordev/vkit-zenith/packages/tools/dbtool"
)

// Only test tooling reads this launcher-provided private path. Child runtime
// environments do not receive it; this is not a feature configuration override.
func ownershipContext(ctx context.Context) context.Context {
	return dbtool.WithOwnershipJournal(ctx, os.Getenv("AUTH_OWNERSHIP_MANIFEST"))
}

func recordOwnership(kind, id string) error {
	return dbtool.RecordOwnership(ownershipContext(context.Background()), kind, id)
}

// The supervisor owns this parent BEFORE Go starts. A hard kill between mkdir
// and journalling therefore cannot leave an undiscoverable external root.
func ownedRoot(prefix string) (string, error) {
	parent := "/tmp/opencode"
	if journal := os.Getenv("AUTH_OWNERSHIP_MANIFEST"); journal != "" {
		parent = filepath.Join(filepath.Dir(journal), "r")
	}
	root, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return "", err
	}
	if err := recordOwnership("root", root); err != nil {
		_ = os.RemoveAll(root)
		return "", err
	}
	return root, nil
}
