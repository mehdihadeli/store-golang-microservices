package migrations

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"go.uber.org/zap"
)

type MigrationParams struct {
	DbName        string `mapstructure:"dbName"`
	VersionTable  string `mapstructure:"versionTable"`
	MigrationsDir string `mapstructure:"migrationsDir"`
	TargetVersion uint   `mapstructure:"targetVersion"`
	SkipMigration bool   `mapstructure:"skipMigration"`
}

func RunMigration(db *sql.DB, p MigrationParams) error {
	d, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: p.VersionTable,
		DatabaseName:    p.DbName,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+p.MigrationsDir, p.DbName, d)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if p.TargetVersion == 0 {
		err = m.Up()
	} else {
		err = m.Migrate(p.TargetVersion)
	}

	if err == migrate.ErrNoChange {
		return nil
	}

	zap.L().Info("migration finished")
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
