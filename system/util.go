package system

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"
	"fmt"
	"net/url"
)

var (
	// Signals that we will handle
	signals = []os.Signal{syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT}
)

func WaitForSignal() os.Signal {
	// Catch signals
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, signals...)

	// Wait for a signal to exit
	signal := <-sigchan
	return signal
}

func SetGlobalTimezoneUTC() error {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

func ParseDBConnURL(args []string) (string, error) {
	var dBConnURL string

	fs := flag.FlagSet{}
	fs.StringVar(&dBConnURL, "db", os.Getenv("DATABASE_URL"), "Postgres connection URL, eg: postgres://user:pass@host:5432/dbname. Must be a valid URL. Defaults to DATABASE_URL")

	err := fs.Parse(args)
	if err != nil {
		return "", err
	}

	// Postgres URLs follow this form: postgres://user:password@host:port/dbname?args
	// Parse them as a URL to ensure they are valid, otherwise return an error.
	_, err = url.Parse(dBConnURL)
	if err != nil {
		return "", fmt.Errorf("the -db or DATABASE_URL url is not valid")
	}

	if dBConnURL == "" {
		return "", fmt.Errorf("missing -db or DATABASE_URL")
	}

	return dBConnURL, nil
}

func ParseDisbursementDateRange(args []string) (*time.Time, *time.Time, error) {
	var ordersFrom string
	var ordersTo string

	fs := flag.FlagSet{}
	fs.StringVar(&ordersFrom, "from", os.Getenv("ORDERS_FROM"), "Start date orders to process, eg: 2021-10-01. Must be a valid Date yyyy-mm-dd. Defaults to ''")
	fs.StringVar(&ordersTo, "to", os.Getenv("ORDERS_TO"), "End date orders to process, eg: 2021-10-01. Must be a valid Date yyyy-mm-dd. Defaults to ''")

	err := fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	// If no date range, ignore
	if ordersFrom == "" || ordersTo == "" {
		return nil, nil, nil
	}

	ordersFromDate, err := time.Parse(time.DateOnly, ordersFrom)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing ORDERS_FROM: %v", err)
	}

	ordersToDate, err := time.Parse(time.DateOnly, ordersTo)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing ORDERS_TO: %v", err)
	}

	return &ordersFromDate, &ordersToDate, nil
}

func FirstDayOfLastMonth(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month()-1, 1, 0, 0, 0, 0, day.Location())
}

func LastDayOfLastMonth(day time.Time) time.Time {
	return FirstDayOfLastMonth(day).AddDate(0, 1, -1)
}
