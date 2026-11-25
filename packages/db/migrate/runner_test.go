package migrate

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
)

type unreadableFS struct{}

func (unreadableFS) Open(string) (fs.File, error) { return nil, fs.ErrPermission }

func TestSourceValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source fstest.MapFS
		bad    bool
	}{
		{"empty", fstest.MapFS{"README.md": {Data: []byte("empty")}}, false},
		{"sorted", fstest.MapFS{"00002_b.sql": {Data: []byte("-- +goose Up\nSELECT 2;\n")}, "00001_a.sql": {Data: []byte("-- +goose Up\nSELECT 1;\n")}}, false},
		{"malformed", fstest.MapFS{"bad.sql": {Data: []byte("-- +goose Up\nSELECT 1;")}}, true},
		{"duplicate", fstest.MapFS{"1_a.sql": {Data: []byte("-- +goose Up\nSELECT 1;")}, "01_b.sql": {Data: []byte("-- +goose Up\nSELECT 2;")}}, true},
		{"missing-up", fstest.MapFS{"1_a.sql": {Data: []byte("SELECT 1;")}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := Source(tc.source)
			if (err != nil) != tc.bad {
				t.Fatalf("unexpected validation result: %v", err)
			}
			if tc.name == "sorted" && (entries[0] != 1 || entries[1] != 2) {
				t.Fatal("unsorted")
			}
		})
	}
	if _, err := Source(nil); err == nil {
		t.Fatal("nil accepted")
	}
	if _, err := Source(unreadableFS{}); err == nil {
		t.Fatal("unreadable source accepted as empty")
	}
	if _, err := Source(os.DirFS("testdata/does-not-exist")); err == nil {
		t.Fatal("missing source accepted as empty")
	}
	if _, err := Source(fstest.MapFS{"1_file.sql": {Mode: os.ModeSymlink, Data: []byte("-- +goose Up\nSELECT 1;\n")}}); err == nil {
		t.Fatal("symlink migration accepted")
	}
}

func TestRequestSafety(t *testing.T) {
	for _, r := range []Request{{Operation: "other"}, {Operation: DownTo, Target: 0}, {Operation: DownTo, Target: -1, ConfirmRollback: true}} {
		if _, err := Run(context.Background(), nil, fstest.MapFS{}, r); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
	r, err := Run(context.Background(), nil, fstest.MapFS{}, Request{Operation: Status})
	if err != nil || r.Version != 0 || len(r.Entries) != 0 {
		t.Fatalf("empty state: %v", err)
	}
}
