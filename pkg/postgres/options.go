package postgres

import "time"

type Option func(*postgres)

func WithConnAttempts(attempts int) Option {
	return func(m *postgres) {
		m.connAttempts = attempts
	}
}

func WithConnTimeout(timeout time.Duration) Option {
	return func(m *postgres) {
		m.connTimeout = timeout
	}
}
