package rcon

import "time"

// Settings contains option to Conn.
type Settings struct {
	dialTimeout time.Duration
	deadline    time.Duration
}

// DefaultSettings provides default deadline settings to Conn.
var DefaultSettings = Settings{
	dialTimeout: DefaultDialTimeout,
	deadline:    DefaultDeadline,
}

// Option allows to inject settings to Settings.
type Option func(s *Settings)

// SetDialTimeout injects dial Timeout to Settings.
func SetDialTimeout(timeout time.Duration) Option {
	return func(s *Settings) {
		s.dialTimeout = timeout
	}
}

// SetDeadline injects read/write Timeout to Settings.
func SetDeadline(timeout time.Duration) Option {
	return func(s *Settings) {
		s.deadline = timeout
	}
}
