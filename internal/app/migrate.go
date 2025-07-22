//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 5
	_defaultTimeout  = time.Second
	_migrationPath   = "db/migrations"
	_dbURL           = "postgres://testuser:testpassword@postgres:5432/escape?sslmode=disable&x-migrations-table=ordersrv_schema_migrations"
)

func init() {
	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)
	for attempts > 0 {
		dir := fmt.Sprintf("file://%s", _migrationPath)
		m, err = migrate.New(dir, _dbURL)
		if err == nil {
			break
		}
		slog.Info("Migrate: postgres is trying to connect", "attempts", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}
	if err != nil {
		slog.Error("Migrate: postgres connect error", "error", err)
		return
	}
	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Migrate: up error", "error", err)
		return
	}
	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("Migrate: no change")
		return
	}
	slog.Info("Migrate: up success")
}
