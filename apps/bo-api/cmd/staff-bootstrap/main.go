package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/bootstrapcli"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

func run(args []string) error {
	flags := flag.NewFlagSet("staff-bootstrap", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "Trusted application root")
	development := flags.Bool("development", false, "Explicit .env opt-in")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("Argumen tidak valid; gunakan --root PATH [--development].")
	}
	trusted, err := config.ApplicationRoot(*root)
	if err != nil {
		return errors.New("Root aplikasi tidak valid.")
	}
	in, err := bootstrapcli.Collect(bootstrapcli.Console{In: os.Stdin, Out: os.Stderr}, os.Stderr)
	if err != nil {
		return err
	}
	opts := config.Options{Root: trusted}
	if *development {
		opts.Environment, err = config.DevelopmentEnvironment(trusted, nil)
		if err != nil {
			return errors.New("Konfigurasi pengembangan tidak tersedia.")
		}
	}
	d, err := config.LoadDatabase(opts)
	if err != nil {
		return errors.New("Konfigurasi database tidak tersedia.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, err := db.Open(ctx, db.Settings{DSN: d.DSN, MaxOpenConnections: d.MaxOpenConnections, ConnectionMaxLifetimeSeconds: d.ConnectionMaxLifetimeSeconds, ConnectionMaxIdleTimeSeconds: d.ConnectionMaxIdleTimeSeconds, ConnectTimeoutSeconds: d.ConnectTimeoutSeconds})
	if err != nil {
		return errors.New("Database tidak tersedia.")
	}
	defer p.Close()
	s, err := staff.NewService(p, time.Now, staff.NewPasswords()).Bootstrap(ctx, in)
	if errors.Is(err, staff.ErrConflict) {
		return errors.New("Bootstrap ditolak: Superadmin pernah dibuat atau email sudah digunakan.")
	}
	if err != nil {
		return errors.New("Bootstrap gagal; tidak ada kredensial yang ditampilkan.")
	}
	_, err = fmt.Fprintln(os.Stdout, "Superadmin dibuat:", s.ID.String())
	return err
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
