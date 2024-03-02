package rcontest

import (
	"net"

	"github.com/gorcon/rcon"
)

// Context represents the context of the current RCON request. It holds request
// and conn objects.
type Context struct {
	server  *Server
	conn    net.Conn
	request *rcon.Packet
}

// Server returns the Server instance.
func (c *Context) Server() *Server {
	return c.server
}

// Conn returns current RCON connection.
func (c *Context) Conn() net.Conn {
	return c.conn
}

// Request returns received *rcon.Packet.
func (c *Context) Request() *rcon.Packet {
	return c.request
}
