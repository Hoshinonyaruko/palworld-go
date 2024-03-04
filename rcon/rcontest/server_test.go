package rcontest_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
)

func TestNewServer(t *testing.T) {
	t.Run("with options", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "password"}),
			rcontest.SetAuthHandler(func(c *rcontest.Context) {
				if c.Request().Body() == c.Server().Settings.Password {
					rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, c.Request().ID, "").WriteTo(c.Conn())
				} else {
					rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(c.Conn())
				}
			}),
			rcontest.SetCommandHandler(func(c *rcontest.Context) {
				rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "Can I help you?").WriteTo(c.Conn())
			}),
		)
		defer server.Close()

		client, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		response, err := client.Execute("Can I help you?")
		if err != nil {
			t.Fatal(err)
		}

		if response != "Can I help you?" {
			t.Errorf("got %q, want Can I help you?", response)
		}
	})

	t.Run("unstarted", func(t *testing.T) {
		server := rcontest.NewUnstartedServer()
		server.Settings.Password = "password"
		server.Settings.AuthResponseDelay = 10 * time.Millisecond
		server.Settings.CommandResponseDelay = 10 * time.Millisecond
		server.SetCommandHandler(func(c *rcontest.Context) {
			rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "I do it all.").WriteTo(c.Conn())
		})
		server.Start()
		defer server.Close()

		client, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		response, err := client.Execute("What do you do?")
		if err != nil {
			t.Fatal(err)
		}

		if response != "I do it all." {
			t.Errorf("got %q, want I do it all.", response)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		server := rcontest.NewServer()
		defer server.Close()

		client, err := rcon.Dial(server.Addr(), "wrong")
		if err != nil {
			defer client.Close()
		}
		if !errors.Is(err, rcon.ErrAuthFailed) {
			t.Fatal(err)
		}
	})

	t.Run("empty handler", func(t *testing.T) {
		server := rcontest.NewServer()
		defer server.Close()

		client, err := rcon.Dial(server.Addr(), "")
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		response, err := client.Execute("whatever")
		if err != nil {
			t.Fatal(err)
		}

		if response != "" {
			t.Errorf("got %q, want empty string", response)
		}
	})
}
