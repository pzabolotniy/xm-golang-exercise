package migration

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pzabolotniy/logging/pkg/logging"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
)

func MigrateUp(ctx context.Context, dbConn *sqlx.DB, migrationConf *config.DB) error {
	logger := logging.FromContext(ctx)
	migrations := &migrate.FileMigrationSource{
		Dir: migrationConf.MigrationDir,
	}

	migrate.SetTable(migrationConf.MigrationTable)
	migrationsApplied, err := migrate.Exec(dbConn.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.WithError(err).Error("migration failed")

		return err
	}
	logger.WithField("migrations_applied", migrationsApplied).Trace("migration succeeded")

	return nil
}
