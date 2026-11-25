package dbtool

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempRoot(t *testing.T) string {
	t.Helper()
	p, e := os.MkdirTemp("/tmp/opencode", "dbtool-")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(p); err != nil {
			t.Errorf("owned temporary directory cleanup failed: %v", err)
		}
	})
	return p
}
func TestPublicationSafety(t *testing.T) {
	root := tempRoot(t)
	out := filepath.Join(root, "generated")
	os.Mkdir(out, 0700)
	os.WriteFile(filepath.Join(out, "README.md"), []byte("handwritten"), 0600)
	generate := func(ctx context.Context, dir string) error {
		return os.WriteFile(filepath.Join(dir, "model.go"), []byte(generatedNotice+"\npackage models\n"), 0600)
	}
	if err := publish(context.Background(), out, generate); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(out, "model.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := publish(context.Background(), out, func(context.Context, string) error { return errors.New("injected generation failure") }); err == nil {
		t.Fatal("failure hidden")
	}
	current, _ := os.ReadFile(filepath.Join(out, "model.go"))
	if string(current) != string(original) {
		t.Fatal("previous output changed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if publish(ctx, out, generate) == nil {
		t.Fatal("cancel ignored")
	}
	os.WriteFile(filepath.Join(out, "user.go"), []byte("package user"), 0600)
	if publish(context.Background(), out, generate) == nil {
		t.Fatal("handwritten overwrite permitted")
	}
	os.Remove(filepath.Join(out, "user.go"))
	os.Symlink(filepath.Join(out, "model.go"), filepath.Join(out, "link.go"))
	if publish(context.Background(), out, generate) == nil {
		t.Fatal("symlink accepted")
	}
	os.Remove(filepath.Join(out, "link.go"))
	if err := publish(context.Background(), out, func(context.Context, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "model.go")); !os.IsNotExist(err) {
		t.Fatal("stale owned output retained")
	}
	readme, _ := os.ReadFile(filepath.Join(out, "README.md"))
	if string(readme) != "handwritten" {
		t.Fatal("README overwritten")
	}
}

func TestPublicationConcurrentAndCancelled(t *testing.T) {
	root := tempRoot(t)
	out := filepath.Join(root, "generated")
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- publish(context.Background(), out, func(context.Context, string) error { close(entered); <-release; return nil })
	}()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	second := make(chan error, 1)
	interleaved := make(chan struct{}, 1)
	go func() {
		second <- publish(ctx, out, func(context.Context, string) error { interleaved <- struct{}{}; return nil })
	}()
	select {
	case <-interleaved:
		t.Error("concurrent generation entered")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	if err := <-second; err != context.Canceled {
		t.Errorf("waiting cancellation lost: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := publish(context.Background(), out, func(ctx context.Context, stage string) error {
		if err := os.WriteFile(filepath.Join(stage, "partial.go"), []byte(generatedNotice+"\npackage models\n"), 0600); err != nil {
			return err
		}
		return context.Canceled
	}); err == nil {
		t.Fatal("cancelled staged work published")
	}
	if _, err := os.Stat(filepath.Join(out, "partial.go")); !os.IsNotExist(err) {
		t.Fatal("partial output published")
	}
	stages, _ := filepath.Glob(filepath.Join(root, ".dbgen-stage-*"))
	if len(stages) != 0 {
		t.Fatal("staging leaked")
	}
}

func TestPublicationRecoveryBackupRefused(t *testing.T) {
	root := tempRoot(t)
	out := filepath.Join(root, "generated")
	backup := out + ".backup"
	if err := os.Mkdir(backup, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backup, "README.md"), []byte("recover me"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publish(context.Background(), out, func(context.Context, string) error { t.Error("generation before recovery inspection"); return nil }); err == nil {
		t.Fatal("backup overwritten")
	}
	data, err := os.ReadFile(filepath.Join(backup, "README.md"))
	if err != nil || string(data) != "recover me" {
		t.Fatal("backup lost")
	}
}
