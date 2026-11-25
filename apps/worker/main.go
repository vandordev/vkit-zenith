package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vandordev/vkit-zenith/packages/config/go/config"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

func main() {
	root := flag.String("root", "", "Trusted application root containing config/config.yaml")
	flag.Parse()
	if err := run(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string) error {
	root, err := config.ApplicationRoot(root)
	if err != nil {
		return errors.New("worker configuration failed")
	}
	options := config.Options{Root: root}
	database, err := config.LoadDatabase(options)
	if err != nil {
		return errors.New("worker configuration failed")
	}
	worker, err := config.LoadWorker(options)
	if err != nil {
		return errors.New("worker configuration failed")
	}
	pool, err := db.Open(context.Background(), db.Settings{
		DSN:                          database.DSN,
		MaxOpenConnections:           database.MaxOpenConnections,
		ConnectionMaxLifetimeSeconds: database.ConnectionMaxLifetimeSeconds,
		ConnectionMaxIdleTimeSeconds: database.ConnectionMaxIdleTimeSeconds,
		ConnectTimeoutSeconds:        database.ConnectTimeoutSeconds,
	})
	if err != nil {
		return errors.New("worker database unavailable")
	}
	defer pool.Close()

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: worker.MaxWorkers}},
		Workers: river.NewWorkers(),
	})
	if err != nil {
		return errors.New("worker configuration failed")
	}
	if err := client.Start(context.Background()); err != nil {
		return errors.New("worker has no registered jobs")
	}
	return errors.New("worker has no registered jobs")
}
