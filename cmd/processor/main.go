package main

import (
	"context"
	"github.com/ildomm/cc_sq_disbursement/database"
	"github.com/ildomm/cc_sq_disbursement/order_process"
	"github.com/ildomm/cc_sq_disbursement/system"
	"log"
	"os"
	"time"
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
	log.SetPrefix("Processor Job: ")
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

	// Check arg for date range
	start, end, err := system.ParseDisbursementDateRange(os.Args[1:])
	if err != nil {
		log.Fatalf("parsing command line: %s", err)
	}

	pipeline := order_process.NewPipeline(ctx, querier)
	interval := 24 * time.Hour // One day interval

	if start != nil && end != nil {
		// Run for each day in the date range
		for current := *start; current.Before(*end); current = current.Add(interval) {
			pipeline.Run(current)
		}
	} else {
		// If no date range, run for yesterday
		pipeline.Run(time.Now().Add(-interval))
	}
}
