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

func TestSetGlobalTimezoneUTC(t *testing.T) {
	err := SetGlobalTimezoneUTC()
	require.NoError(t, err)

	// Check if time.Local is set to UTC
	require.Equal(t, time.UTC, time.Local)

	// Optional: Test some time-related functions
	now := time.Now()
	require.Equal(t, "UTC", now.Location().String())
}

func TestValidDbURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/dbname")
	defer os.Unsetenv("DATABASE_URL")

	url, err := ParseDBConnURL([]string{})
	require.NoError(t, err)
	require.Equal(t, "postgres://user:pass@host:5432/dbname", url)
}

func TestFirstDayOfLastMonth(t *testing.T) {
	tests := []struct {
		day      time.Time
		expected time.Time
	}{
		{time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC), time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		result := FirstDayOfLastMonth(test.day)
		require.Equal(t, test.expected, result)
	}
}

func TestLastDayOfLastMonth(t *testing.T) {
	tests := []struct {
		day      time.Time
		expected time.Time
	}{
		{time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC), time.Date(2023, 4, 30, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		result := LastDayOfLastMonth(test.day)
		require.Equal(t, test.expected, result)
	}
}

func TestParseDisbursementDateRangeInvalidFrom(t *testing.T) {
	invalidFromDate := "invalid-date"
	args := []string{"-from", invalidFromDate, "-to", "2021-10-01"}

	from, to, err := ParseDisbursementDateRange(args)
	require.Error(t, err)
	require.Nil(t, from)
	require.Nil(t, to)
	require.Contains(t, err.Error(), "error parsing ORDERS_FROM")
}

func TestParseDisbursementDateRangeInvalidTo(t *testing.T) {
	invalidToDate := "invalid-date"
	args := []string{"-from", "2021-10-01", "-to", invalidToDate}

	from, to, err := ParseDisbursementDateRange(args)
	require.Error(t, err)
	require.Nil(t, from)
	require.Nil(t, to)
	require.Contains(t, err.Error(), "error parsing ORDERS_TO")
}
