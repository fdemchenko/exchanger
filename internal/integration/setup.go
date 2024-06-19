package integration

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	LogOccurance          = 2
	ContainerStartTimeout = 10 * time.Second
)

func createTestDBContainer() (*postgres.PostgresContainer, error) {
	ctx := context.Background()

	dbName := "test"
	dbUser := "testuser"
	dbPassword := "postgres"

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(LogOccurance).
				WithStartupTimeout(ContainerStartTimeout)),
	)
	if err != nil {
		return nil, err
	}
	_, _, err = postgresContainer.Exec(ctx, []string{"psql", "-U", dbUser,
		"-d", dbName,
		"-c", "CREATE TABLE emails (id SERIAL PRIMARY KEY, email TEXT UNIQUE NOT NULL)"})
	if err != nil {
		return nil, err
	}
	err = postgresContainer.Snapshot(ctx, postgres.WithSnapshotName("test-db"))
	if err != nil {
		return nil, err
	}

	return postgresContainer, nil
}
