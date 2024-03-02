// Package rcontest contains RCON server for RCON client testing.
package rcontest

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorcon/rcon"
)

// Server is an RCON server listening on a system-chosen port on the
// local loopback interface, for use in end-to-end RCON tests.
type Server struct {
	Settings       Settings
	Listener       net.Listener
	addr           string
	authHandler    HandlerFunc
	commandHandler HandlerFunc
	connections    map[net.Conn]struct{}
	quit           chan bool
	wg             sync.WaitGroup
	mu             sync.Mutex
	closed         bool
}

// Settings contains configuration for RCON Server.
type Settings struct {
	Password             string
	AuthResponseDelay    time.Duration
	CommandResponseDelay time.Duration
}

// HandlerFunc defines a function to serve RCON requests.
type HandlerFunc func(c *Context)

// AuthHandler checks authorisation data and responses with
// SERVERDATA_AUTH_RESPONSE packet.
func AuthHandler(c *Context) {
	if c.Request().Body() == c.Server().Settings.Password {
		// First write SERVERDATA_RESPONSE_VALUE packet with empty body.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "").WriteTo(c.Conn())

		// Than write SERVERDATA_AUTH_RESPONSE packet to allow authHandler success.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, rcon.SERVERDATA_AUTH_ID, "").WriteTo(c.Conn())
	} else {
		// If authentication was failed, the ID must be assigned to -1.
		_, _ = rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(c.Conn())
	}
}

// EmptyHandler responses with empty body. Is used when start RCON Server with nil
// commandHandler.
func EmptyHandler(c *Context) {
	_, _ = rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "").WriteTo(c.Conn())
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("rcontest: failed to listen on a port: %v", err))
	}

	return l
}

// NewServer returns a running RCON Server or nil if an error occurred.
// The caller should call Close when finished, to shut it down.
func NewServer(options ...Option) *Server {
	server := NewUnstartedServer(options...)
	server.Start()

	return server
}

// NewUnstartedServer returns a new Server but doesn't start it.
// After changing its configuration, the caller should call Start.
// The caller should call Close when finished, to shut it down.
func NewUnstartedServer(options ...Option) *Server {
	server := Server{
		Listener:       newLocalListener(),
		authHandler:    AuthHandler,
		commandHandler: EmptyHandler,
		connections:    make(map[net.Conn]struct{}),
		quit:           make(chan bool),
	}

	for _, option := range options {
		option(&server)
	}

	return &server
}

// SetAuthHandler injects HandlerFunc with authorisation data checking.
func (s *Server) SetAuthHandler(handler HandlerFunc) {
	s.authHandler = handler
}

// SetCommandHandler injects HandlerFunc with commands processing.
func (s *Server) SetCommandHandler(handler HandlerFunc) {
	s.commandHandler = handler
}

// Start starts a server from NewUnstartedServer.
func (s *Server) Start() {
	if s.addr != "" {
		panic("server already started")
	}

	s.addr = s.Listener.Addr().String()
	s.goServe()
}

// Close shuts down the Server.
func (s *Server) Close() {
	if s.closed {
		return
	}

	s.closed = true
	close(s.quit)
	s.Listener.Close()

	// Waiting for server connections.
	s.wg.Wait()

	s.mu.Lock()
	for c := range s.connections {
		// Force-close any connections.
		s.closeConn(c)
	}
	s.mu.Unlock()
}

// Addr returns IPv4 string Server address.
func (s *Server) Addr() string {
	return s.addr
}

// NewContext returns a Context instance.
func (s *Server) NewContext(conn net.Conn) (*Context, error) {
	ctx := Context{server: s, conn: conn, request: &rcon.Packet{}}

	if _, err := ctx.request.ReadFrom(conn); err != nil {
		return &ctx, fmt.Errorf("rcontest: %w", err)
	}

	return &ctx, nil
}

// serve handles incoming requests until a stop signal is given with Close.
func (s *Server) serve() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.isRunning() {
				panic(fmt.Errorf("rcontest: %w", err))
			}

			return
		}

		s.wg.Add(1)

		go s.handle(conn)
	}
}

// serve calls serve in goroutine.
func (s *Server) goServe() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		s.serve()
	}()
}

// handle handles incoming client conn.
func (s *Server) handle(conn net.Conn) {
	s.mu.Lock()
	s.connections[conn] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.closeConn(conn)
		s.wg.Done()
	}()

	for {
		ctx, err := s.NewContext(conn)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(fmt.Errorf("failed read request: %w", err))
			}

			return
		}

		switch ctx.Request().Type {
		case rcon.SERVERDATA_AUTH:
			if s.Settings.AuthResponseDelay != 0 {
				time.Sleep(s.Settings.AuthResponseDelay)
			}

			s.authHandler(ctx)
		case rcon.SERVERDATA_EXECCOMMAND:
			if s.Settings.CommandResponseDelay != 0 {
				time.Sleep(s.Settings.CommandResponseDelay)
			}

			s.commandHandler(ctx)
		}
	}
}

// isRunning returns true if Server is running and false if is not.
func (s *Server) isRunning() bool {
	select {
	case <-s.quit:
		return false
	default:
		return true
	}
}

// closeConn closes a client conn and removes it from connections map.
func (s *Server) closeConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := conn.Close(); err != nil {
		panic(fmt.Errorf("close conn error: %w", err))
	}

	delete(s.connections, conn)
}
