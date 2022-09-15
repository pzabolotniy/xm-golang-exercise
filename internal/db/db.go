package db

import (
	"context"

	// Load PostgreSQL driver.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
)

func Connect(ctx context.Context, dbConf *config.DB) (*sqlx.DB, error) {
	logger := logging.FromContext(ctx)
	connString := dbConf.ConnString
	conn, err := sqlx.Connect("pgx", connString)
	if err != nil {
		logger.WithError(err).Error("connect failed")

		return nil, err
	}

	return conn, nil
}

func Disconnect(dbConn *sqlx.DB) error {
	return dbConn.Close()
}
