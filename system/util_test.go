package system

import (
	"github.com/stretchr/testify/require"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"testing/quick"
	"time"
)

func TestWaitForSignal(t *testing.T) {
	// Use a buffered channel to avoid blocking the sender goroutine
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT)

	go func() {
		// Simulate sending a signal after a delay
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Add a delay to allow the signal handler to execute
	time.Sleep(100 * time.Millisecond)

	signalReceived := WaitForSignal()

	if signalReceived != syscall.SIGINT {
		t.Errorf("Expected signal SIGINT, got %v", signalReceived)
	}
}

func TestMissingDbURL(t *testing.T) {
	_, err := ParseDBConnURL([]string{})
	if err == nil || err.Error() != "missing -db or DATABASE_URL" {
		t.Fatalf("Wrong error, got %v", err)
	}
}

func TestInvalidDbURL(t *testing.T) {
	_, err := ParseDBConnURL([]string{
		"-db",
		"postgres://user:pass@host:port-not-a-number/dbname2"})
	if err == nil || err.Error() != "the -db or DATABASE_URL url is not valid" {
		t.Fatalf("Wrong error, got %v", err)
	}
}

func TestMissingDisbursementDateRange(t *testing.T) {
	_, _, err := ParseDisbursementDateRange([]string{})
	require.NoError(t, err)
}

func TestParseDisbursementDateRangeInvalidParameters(t *testing.T) {
	// Define the function for property-based testing
	f := func(from, to string) bool {
		// Create invalid parameter combinations
		args := []string{"-ORDERS_FROM", from, "-ORDERS_TO", to}

		// Call the ParseDisbursementDateRange method with invalid parameters
		_, _, err := ParseDisbursementDateRange(args)

		// Check if error is expected for invalid parameters
		return err != nil
	}

	// Run the property-based testing
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
