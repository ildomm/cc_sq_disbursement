package main

import (
	"context"
	"github.com/ildomm/cc_sq_disbursement/database"
	"github.com/ildomm/cc_sq_disbursement/order_load"
	"github.com/ildomm/cc_sq_disbursement/system"
	"log"
	"os"
)

var (
	semVer = "unknown" // Populated with semantic version at build time
)

func main() {
	// Create an overarching context which we can use to safely cancel
	// all goroutines when we receive a signal to terminate.
	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("Loader Job: ")
	log.Printf("starting job, Version %s", semVer)

	// Set the timezone to UTC
	system.SetGlobalTimezoneUTC() //nolint:all

	// Parse the command line options
	dBConnURL, err := system.ParseDBConnURL(os.Args[1:])
	if err != nil {
		log.Fatalf("parsing command line: %s", err)
	}

	// Set up the database connection and run migrations
	log.Printf("connecting to database")
	querier, err := database.NewPostgresQuerier(
		ctx,
		dBConnURL,
	)
	if err != nil {
		log.Fatalf("error connecting to the database: %s", err)
	}
	defer querier.Close()

	// Start pipeline as a goroutine
	go order_load.NewPipeline(ctx, querier).Run()

	log.Printf("caught signal, terminating. Signal %s", system.WaitForSignal().String())
}
