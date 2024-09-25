package migrate

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
)

var (
	ErrMigrateVersion       = errors.New("error on migrate version")
	ErrDatabaseConnection   = errors.New("database connection is nil")
	ErrUnableToCreateDriver = errors.New("unable to create driver instance")
)

type (
	Migrate interface {
		Execute() error
	}

	migration struct {
		migrate *migrate.Migrate
	}
)

func (m *migration) Execute() error {
	_, _, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Println(err)
		return ErrMigrateVersion
	}

	err = m.migrate.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println(err)
		return nil
	}

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
