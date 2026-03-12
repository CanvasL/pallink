package postgres

// Config is used for postgres connection settings.
type Config struct {
	DSN                      string
	MaxConns                 int32
	MinConns                 int32
	MaxConnLifetimeSeconds   int64
	MaxConnIdleSeconds       int64
	HealthCheckPeriodSeconds int64
}
