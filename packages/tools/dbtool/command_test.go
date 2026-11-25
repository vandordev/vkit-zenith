package dbtool

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestMigrationCommandParsing(t *testing.T) {
	for _, args := range [][]string{nil, {"--root", "/missing", "--operation", "other"}, {"--root", "/missing", "--operation", "down-to"}, {"--root", "/missing", "--operation", "down-to", "--target", "0"}, {"--root", "/missing", "--operation", "down-to", "--confirm-rollback"}, {"--root", "/missing", "--operation", "down-to", "--target", "-1", "--confirm-rollback"}, {"--root", "/missing", "--operation", "up", "--unknown", "sentinel-secret"}} {
		var out bytes.Buffer
		err := MigrationCommand(context.Background(), args, &out)
		if err == nil {
			t.Fatal("invalid command accepted")
		}
		if strings.Contains(err.Error()+out.String(), "sentinel-secret") {
			t.Fatal("argument secret leaked")
		}
	}
	var out bytes.Buffer
	if err := MigrationCommand(context.Background(), []string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "development") {
		t.Fatal("help missing explicit development")
	}
}

func TestMigrationCommandDevelopmentOptIn(t *testing.T) {
	root := tempRoot(t)
	// Without --development no env adapter is called: failure is missing YAML,
	// not missing .env. With explicit opt-in only the temporary root is inspected.
	var out bytes.Buffer
	err := MigrationCommand(context.Background(), []string{"--root", root, "--operation", "status"}, &out)
	if err == nil || strings.Contains(err.Error(), "development") {
		t.Fatal("implicit development env read")
	}
	err = MigrationCommand(context.Background(), []string{"--root", root, "--operation", "status", "--development"}, &out)
	if err == nil || !strings.Contains(err.Error(), "development") {
		t.Fatal("explicit development env adapter not used")
	}
}

func TestMigrationCommandInvalidRoot(t *testing.T) {
	for _, root := range []string{"", tempRoot(t) + "/missing"} {
		var out bytes.Buffer
		if err := MigrationCommand(context.Background(), []string{"--root", root, "--operation", "status"}, &out); err == nil {
			t.Fatal("invalid root accepted")
		}
	}
}

func TestGenerationCommandParsing(t *testing.T) {
	for _, args := range [][]string{nil, {"--root", tempRoot(t) + "/missing"}, {"--dsn", "sentinel-secret"}, {"unexpected"}} {
		var out bytes.Buffer
		err := GenerationCommand(context.Background(), args, &out)
		if err == nil {
			t.Fatal("invalid generation command accepted")
		}
		if strings.Contains(err.Error()+out.String(), "sentinel-secret") {
			t.Fatal("argument secret leaked")
		}
	}
	var out bytes.Buffer
	if err := GenerationCommand(context.Background(), []string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "offline") {
		t.Fatal("help missing offline contract")
	}
}
