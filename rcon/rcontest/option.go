package rcontest

// Option allows to inject Settings to Server.
type Option func(s *Server)

// SetSettings injects configuration for RCON Server.
func SetSettings(settings Settings) Option {
	return func(s *Server) {
		s.Settings = settings
	}
}

// SetAuthHandler injects HandlerFunc with authorisation data checking.
func SetAuthHandler(handler HandlerFunc) Option {
	return func(s *Server) {
		s.SetAuthHandler(handler)
	}
}

// SetCommandHandler injects HandlerFunc with commands processing.
func SetCommandHandler(handler HandlerFunc) Option {
	return func(s *Server) {
		s.SetCommandHandler(handler)
	}
}
