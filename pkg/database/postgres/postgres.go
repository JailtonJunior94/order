package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jailtonjunior94/order/configs"

	_ "github.com/lib/pq"
)

var (
	ErrSQLOpenConn = errors.New("unable to open connection with SQL database")
)

func NewPostgresDatabase(config *configs.Config) (*sql.DB, error) {
	sqlDB, err := sql.Open(config.DBConfig.Driver, dsn(config))
	if err != nil {
		return nil, ErrSQLOpenConn
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, ErrSQLOpenConn
	}

	// Configure connection pool for production
	maxIdleConns := config.DBConfig.DBMaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10 // default
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// Set max open connections to prevent resource exhaustion
	// Default: 2x idle connections (good practice)
	maxOpenConns := maxIdleConns * 2
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// Set connection lifetime to prevent stale connections
	sqlDB.SetConnMaxLifetime(30 * 60 * 1000000000) // 30 minutes
	sqlDB.SetConnMaxIdleTime(10 * 60 * 1000000000) // 10 minutes

	return sqlDB, nil
}

func dsn(config *configs.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBConfig.Host,
		config.DBConfig.Port,
		config.DBConfig.User,
		config.DBConfig.Password,
		config.DBConfig.Name,
	)
}
