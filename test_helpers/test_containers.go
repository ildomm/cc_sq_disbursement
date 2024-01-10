package test_helpers

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	instance testcontainers.Container
}

func NewTestDatabase(t *testing.T) *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13.3",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return &TestDatabase{
		instance: postgres,
	}
}

func (db *TestDatabase) Port(t *testing.T) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := db.instance.MappedPort(ctx, "5432")
	require.NoError(t, err)
	return p.Int()
}

func (db *TestDatabase) ConnectionString(t *testing.T) string {
	name := "postgres"
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s", db.Port(t), name)
}

func (db *TestDatabase) Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	require.NoError(t, db.instance.Terminate(ctx))
}

func (db *TestDatabase) IsRunning(t *testing.T) bool {
	return db.instance.IsRunning()
}
