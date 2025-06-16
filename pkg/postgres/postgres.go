package postgres

import (
	"database/sql"
	"time"
)

const (
	_defaultConnAttempts = 3
	_defaultConnTimeout  = time.Second
)

type DBConnString string

type postgres struct {
	connAttempts int
	connTimeout  time.Duration

	db *sql.DB
}

var _ DBEngine = (*postgres)(nil)

func New(url DBConnString) (DBEngine, error) {
	pg := &postgres{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	var err error
	for pg.connAttempts > 0 {
		pg.db, err = sql.Open("postgres", string(url))
		if err == nil {
			break
		}
		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	return pg, nil
}

func (m *postgres) Configure(opts ...Option) DBEngine {
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *postgres) GetDB() *sql.DB {
	return m.db
}

func (m *postgres) Close() {
	if m.db != nil {
		m.db.Close()
	}
}
