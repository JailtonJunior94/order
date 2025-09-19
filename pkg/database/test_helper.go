package database

import (
	"context"
	"database/sql"
	"log"
	"testing"

	migration "github.com/jailtonjunior94/order/pkg/database/migrate"
	cockroachdbContainer "github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

type CockroachDBContainer struct {
	container *cockroachdbContainer.CockroachDBContainer
}

func SetupCockroachDB(ctx context.Context, t testing.TB, dbName, dbUserName, dbPassword string) *CockroachDBContainer {
	t.Helper()

	cockroachdb, err := cockroachdbContainer.Run(
		ctx,
		"cockroachdb/cockroach:v23.2.4",
		cockroachdbContainer.WithDatabase(dbName),
		cockroachdbContainer.WithNoClusterDefaults(),
	)

	if err != nil {
		t.Errorf("error creating container %s", err.Error())
	}

	state, err := cockroachdb.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		t.Errorf("error creating container %s", err.Error())
	}

	if !state.Running {
		t.Errorf("container is not running")
	}

	return &CockroachDBContainer{container: cockroachdb}
}

func RunMigration(t testing.TB, db *sql.DB, dbName, migratePath string) {
	t.Helper()

	migrate, err := migration.NewMigrateCockroachDB(db, migratePath, dbName)
	if err != nil {
		log.Fatalf("error creating migration %s", err.Error())
	}

	if err = migrate.Execute(); err != nil {
		log.Fatalf("error running migration %s", err.Error())
	}
}

func (te *CockroachDBContainer) Terminate(ctx context.Context, t testing.TB) {
	t.Helper()
	if err := te.container.Terminate(ctx); err != nil {
		t.Errorf("error terminating container %s", err.Error())
	}
}

func (te *CockroachDBContainer) ConnectionString(t testing.TB) string {
	t.Helper()
	connStr, err := te.container.ConnectionString(context.Background())
	if err != nil {
		t.Errorf("error getting connection string %s", err.Error())
	}
	return connStr
}
