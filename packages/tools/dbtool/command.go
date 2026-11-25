package dbtool

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
)

// MigrationCommand is an explicit operational adapter, never a verification
// shortcut. Its availability is not authorization to migrate existing databases.
func MigrationCommand(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("db-migrate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "trusted application root")
	development := flags.Bool("development", false, "explicitly load root .env (injected env wins)")
	op := flags.String("operation", "", "status|version|up|down-to")
	target := flags.String("target", "", "required nonnegative rollback version")
	confirm := flags.Bool("confirm-rollback", false, "explicitly confirm destructive fixture/operator rollback")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(out, "db-migrate --root PATH --operation status|version|up|down-to [--target VERSION --confirm-rollback] [--development]")
			return nil
		}
		return errors.New("migration command: invalid arguments")
	}
	if flags.NArg() != 0 {
		return errors.New("migration command: positional arguments forbidden")
	}
	r := migrate.Request{Operation: migrate.Operation(*op), ConfirmRollback: *confirm}
	if r.Operation != migrate.Status && r.Operation != migrate.Version && r.Operation != migrate.Up && r.Operation != migrate.DownTo {
		return errors.New("migration command: explicit supported operation required")
	}
	if r.Operation == migrate.DownTo {
		v, err := strconv.ParseInt(*target, 10, 64)
		if err != nil || v < 0 || !*confirm {
			return errors.New("migration command: down-to requires target and confirmation")
		}
		r.Target = v
	} else if *target != "" || *confirm {
		return errors.New("migration command: rollback flags require down-to")
	}
	trusted, err := trustedRoot(*root)
	if err != nil {
		return err
	}
	environment := config.Environment()
	if *development {
		environment, err = config.DevelopmentEnvironment(trusted, environment)
		if err != nil {
			return errors.New("migration command: development environment unavailable")
		}
	}
	c, err := config.LoadDatabase(config.Options{Root: trusted, Environment: environment})
	if err != nil {
		return errors.New("migration command: database configuration invalid or unavailable")
	}
	pool, err := migrate.Open(ctx, db.Settings{DSN: c.DSN, MaxOpenConnections: c.MaxOpenConnections, ConnectionMaxLifetimeSeconds: c.ConnectionMaxLifetimeSeconds, ConnectionMaxIdleTimeSeconds: c.ConnectionMaxIdleTimeSeconds, ConnectTimeoutSeconds: c.ConnectTimeoutSeconds})
	if err != nil {
		return err
	}
	result, runErr := migrate.Run(ctx, pool, os.DirFS(filepath.Join(trusted, "packages/db/migrations")), r)
	closeErr := pool.Close()
	if closeErr != nil {
		runErr = errors.Join(runErr, errors.New("migration command: pool close failed"))
	}
	if runErr != nil {
		return runErr
	}
	fmt.Fprintf(out, "version: %d\n", result.Version)
	for _, entry := range result.Entries {
		fmt.Fprintf(out, "%d applied=%t\n", entry.Version, entry.Applied)
	}
	return nil
}

func trustedRoot(root string) (string, error) {
	if root == "" {
		return "", errors.New("database command: explicit trusted --root required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", errors.New("database command: invalid root")
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", errors.New("database command: root directory unavailable")
	}
	return abs, nil
}

// GenerationCommand fixes migration/output paths under the trusted checkout and
// never loads runtime configuration, .env or a user-supplied database DSN.
func GenerationCommand(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("db-gen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "trusted repository root")
	owner := flags.String("owner", "tools", "reviewed owner: tools|application|web-api|bo-api")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(out, "db-gen --root PATH (offline pinned sqlc; no database or Docker)")
			return nil
		}
		return errors.New("generation command: invalid arguments")
	}
	if flags.NArg() != 0 {
		return errors.New("generation command: positional arguments forbidden")
	}
	trusted, err := trustedRoot(*root)
	if err != nil {
		return err
	}
	paths := map[string]string{"tools": "packages/tools/sqlc.yaml", "application": "packages/application/sqlc.yaml", "web-api": "apps/web-api/sqlc.yaml", "bo-api": "apps/bo-api/sqlc.yaml"}
	path, ok := paths[*owner]
	if !ok {
		return errors.New("generation command: unsupported owner")
	}
	return Generate(ctx, filepath.Join(trusted, path))
}
