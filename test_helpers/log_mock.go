package test_helpers

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// captureLogs is a type that implements the io.Writer interface
// to capture log output.
type captureLogs struct {
	logs *[]string
}

func NewLogMocker() *captureLogs {
	var logs []string
	return &captureLogs{&logs}
}

func (c *captureLogs) Write(p []byte) (n int, err error) {
	*c.logs = append(*c.logs, string(p))
	return len(p), nil
}

func (c *captureLogs) String() string {
	return strings.Join(*c.logs, "")
}

func (c *captureLogs) AssertContains(t *testing.T, match string) {
	assert.Contains(t, c.String(), match)
}

func (c *captureLogs) AssertNotContains(t *testing.T, match string) {
	assert.NotContains(t, c.String(), match)
}
