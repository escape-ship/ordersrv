package postgres

import "database/sql"

type DBEngine interface {
	Configure(opts ...Option) DBEngine
	GetDB() *sql.DB
	Close()
}
