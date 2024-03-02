package rcon_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
)

func authHandler(c *rcontest.Context) {
	switch c.Request().Body() {
	case "invalid packet type":
		rcon.NewPacket(42, c.Request().ID, "").WriteTo(c.Conn())
	case "another":
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, 42, "").WriteTo(c.Conn())
	case "makeslice":
		size := int32(len([]byte("")))

		buffer := bytes.NewBuffer(make([]byte, 0, size+4))

		_ = binary.Write(buffer, binary.LittleEndian, size)
		_ = binary.Write(buffer, binary.LittleEndian, c.Request().ID)
		_ = binary.Write(buffer, binary.LittleEndian, rcon.SERVERDATA_RESPONSE_VALUE)

		buffer.WriteTo(c.Conn())
	case c.Server().Settings.Password:
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "").WriteTo(c.Conn())
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, c.Request().ID, "").WriteTo(c.Conn())
	default:
		rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00})).WriteTo(c.Conn())
	}
}

func commandHandler(c *rcontest.Context) {
	writeWithInvalidPadding := func(conn io.Writer, packet *rcon.Packet) {
		buffer := bytes.NewBuffer(make([]byte, 0, packet.Size+4))

		binary.Write(buffer, binary.LittleEndian, packet.Size)
		binary.Write(buffer, binary.LittleEndian, packet.ID)
		binary.Write(buffer, binary.LittleEndian, packet.Type)

		// Write command body, null terminated ASCII string and an empty ASCIIZ string.
		// Second padding byte is incorrect.
		buffer.Write(append([]byte(packet.Body()), 0x00, 0x01))

		buffer.WriteTo(conn)
	}

	switch c.Request().Body() {
	case "help":
		responseBody := "lorem ipsum dolor sit amet"
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, responseBody).WriteTo(c.Conn())
	case "rust":
		// Write specific Rust package.
		rcon.NewPacket(4, c.Request().ID, "").WriteTo(c.Conn())

		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, -1, c.Request().Body()).WriteTo(c.Conn())
	case "padding":
		writeWithInvalidPadding(c.Conn(), rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, ""))
	case "another":
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, 42, "").WriteTo(c.Conn())
	default:
		rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "unknown command").WriteTo(c.Conn())
	}
}

func TestDial(t *testing.T) {
	server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password"}))
	defer server.Close()

	t.Run("connection refused", func(t *testing.T) {
		wantErrContains := "connect: connection refused"

		_, err := rcon.Dial("127.0.0.2:12345", "password")
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("connection timeout", func(t *testing.T) {
		server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password", AuthResponseDelay: 6 * time.Second}))
		defer server.Close()

		wantErrContains := "i/o timeout"

		_, err := rcon.Dial(server.Addr(), "", rcon.SetDialTimeout(5*time.Second))
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		_, err := rcon.Dial(server.Addr(), "wrong")
		if !errors.Is(err, rcon.ErrAuthFailed) {
			t.Errorf("got err %q, want %q", err, rcon.ErrAuthFailed)
		}
	})

	t.Run("invalid packet type", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "password"}),
			rcontest.SetAuthHandler(authHandler),
		)
		defer server.Close()

		_, err := rcon.Dial(server.Addr(), "invalid packet type")
		if !errors.Is(err, rcon.ErrInvalidAuthResponse) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidAuthResponse)
		}
	})

	t.Run("invalid response id", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "password"}),
			rcontest.SetAuthHandler(authHandler),
		)
		defer server.Close()

		_, err := rcon.Dial(server.Addr(), "another")
		if !errors.Is(err, rcon.ErrInvalidPacketID) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidPacketID)
		}
	})

	t.Run("makeslice", func(t *testing.T) {
		server := rcontest.NewServer(
			rcontest.SetSettings(rcontest.Settings{Password: "makeslice"}),
			rcontest.SetAuthHandler(authHandler),
		)
		defer server.Close()

		_, err := rcon.Dial(server.Addr(), "makeslice")
		if !errors.Is(err, rcon.ErrAuthNotRCON) {
			t.Errorf("got err %q, want %q", err, rcon.ErrAuthNotRCON)
		}
	})

	t.Run("auth success", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Errorf("got err %q, want %v", err, nil)
			return
		}

		conn.Close()
	})
}

func TestConn_Execute(t *testing.T) {
	server := rcontest.NewUnstartedServer()
	server.Settings.Password = "password"
	server.SetCommandHandler(commandHandler)
	server.Start()
	defer server.Close()

	t.Run("incorrect command", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("")
		if !errors.Is(err, rcon.ErrCommandEmpty) {
			t.Errorf("got err %q, want %q", err, rcon.ErrCommandEmpty)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}

		result, err = conn.Execute(string(make([]byte, 1001)))
		if !errors.Is(err, rcon.ErrCommandTooLong) {
			t.Errorf("got err %q, want %q", err, rcon.ErrCommandTooLong)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("closed network connection 1", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(0))
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		conn.Close()

		result, err := conn.Execute("help")
		wantErrMsg := fmt.Sprintf("write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr())
		if err == nil || err.Error() != wantErrMsg {
			t.Errorf("got err %q, want to contain %q", err, wantErrMsg)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("closed network connection 2", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		conn.Close()

		result, err := conn.Execute("help")
		wantErrMsg := fmt.Sprintf("rcon: set tcp %s: use of closed network connection", conn.LocalAddr())
		if err == nil || err.Error() != wantErrMsg {
			t.Errorf("got err %q, want to contain %q", err, wantErrMsg)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("read deadline", func(t *testing.T) {
		server := rcontest.NewServer(rcontest.SetSettings(rcontest.Settings{Password: "password", CommandResponseDelay: 2 * time.Second}))
		defer server.Close()

		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(1*time.Second))
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("deadline")
		wantErrMsg := fmt.Sprintf("rcon: read packet size: read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr())
		if err == nil || err.Error() != wantErrMsg {
			t.Errorf("got err %q, want to contain %q", err, wantErrMsg)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("invalid padding", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("padding")
		if !errors.Is(err, rcon.ErrInvalidPacketPadding) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidPacketPadding)
		}

		if len(result) != 2 {
			t.Fatalf("got result len %d, want %d", len(result), 2)
		}
	})

	t.Run("invalid response id", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("another")
		if !errors.Is(err, rcon.ErrInvalidPacketID) {
			t.Errorf("got err %q, want %q", err, rcon.ErrInvalidPacketID)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("success help command", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("help")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}

		resultWant := "lorem ipsum dolor sit amet"
		if result != resultWant {
			t.Fatalf("got result %q, want %q", result, resultWant)
		}
	})

	t.Run("rust workaround", func(t *testing.T) {
		conn, err := rcon.Dial(server.Addr(), "password", rcon.SetDeadline(1*time.Second))
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("rust")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}

		resultWant := "rust"
		if result != resultWant {
			t.Fatalf("got result %q, want %q", result, resultWant)
		}
	})

	if run := getVar("TEST_PZ_SERVER", "false"); run == "true" {
		addr := getVar("TEST_PZ_SERVER_ADDR", "127.0.0.1:16260")
		password := getVar("TEST_PZ_SERVER_PASSWORD", "docker")

		t.Run("pz server", func(t *testing.T) {
			needle := func() string {
				n := `List of server commands :
* additem : Give an item to a player. If no username is given then you will receive the item yourself. Count is optional. Use: /additem "username" "module.item" count. Example: /additem "rj" Base.Axe 5
* adduser : Use this command to add a new user to a whitelisted server. Use: /adduser "username" "password"
* addvehicle : Spawn a vehicle. Use: /addvehicle "script" "user or x,y,z", ex /addvehicle "Base.VanAmbulance" "rj"
* addxp : Give XP to a player. Use /addxp "playername" perkname=xp. Example /addxp "rj" Woodwork=2
* alarm : Sound a building alarm at the Admin's position. (Must be in a room)
* banid : Ban a SteamID. Use /banid SteamID
* banuser : Ban a user. Add a -ip to also ban the IP. Add a -r "reason" to specify a reason for the ban. Use: /banuser "username" -ip -r "reason". For example: /banuser "rj" -ip -r "spawn kill"
* changeoption : Change a server option. Use: /changeoption optionName "newValue"
* chopper : Place a helicopter event on a random player
* createhorde : Spawn a horde near a player. Use : /createhorde count "username". Example /createhorde 150 "rj" Username is optional except from the server console. With no username the horde will be created around you
* createhorde2 : UI_ServerOptionDesc_CreateHorde2
* godmod : Make a player invincible. If no username is set, then you will become invincible yourself. Use: /godmode "username" -value, ex /godmode "rj" -true (could be -false)
* gunshot : Place a gunshot sound on a random player
* help : Help
* invisible : Make a player invisible to zombies. If no username is set then you will become invisible yourself. Use: /invisible "username" -value, ex /invisible "rj" -true (could be -false)
* kick : Kick a user. Add a -r "reason" to specify a reason for the kick. Use: /kickuser "username" -r "reason"
* lightning : Use /lightning "username", username is optional except from the server console
* noclip : Makes a player pass through walls and structures. Toggles with no value. Use: /noclip "username" -value. Example /noclip "rj" -true (could be -false)
* players : List all connected players
* quit : Save and quit the server
* releasesafehouse : Release a safehouse you own. Use /releasesafehouse
* reloadlua : Reload a Lua script on the server. Use /reloadlua "filename"
* reloadoptions : Reload server options (ServerOptions.ini) and send to clients
* removeuserfromwhitelist : Remove a user from the whitelist. Use: /removeuserfromwhitelist "username"
* removezombies : UI_ServerOptionDesc_RemoveZombies
* replay : Record and play replay for moving player. Use /replay "playername" -record|-play|-stop filename. Example: /replay user1 -record stadion.bin
* save : Save the current world
* servermsg : Broadcast a message to all connected players. Use: /servermsg "My Message"
* setaccesslevel : Set access level of a player. Current levels: Admin, Moderator, Overseer, GM, Observer. Use /setaccesslevel "username" "accesslevel". Example /setaccesslevel "rj" "moderator"
* showoptions : Show the list of current server options and values.
* startrain : Starts raining on the server. Use /startrain "intensity", optional intensity is from 1 to 100
* startstorm : Starts a storm on the server. Use /startstorm "duration", optional duration is in game hours
* stoprain : Stop raining on the server
* stopweather : Stop weather on the server
* teleport : Teleport to a player. Once teleported, wait for the map to appear. Use /teleport "playername" or /teleport "player1" "player2". Example /teleport "rj" or /teleport "rj" "toUser"
* teleportto : Teleport to coordinates. Use /teleportto x,y,z. Example /teleportto 10000,11000,0
* thunder : Use /thunder "username", username is optional except from the server console
* unbanid : Unban a SteamID. Use /unbanid SteamID
* unbanuser : Unban a player. Use /unbanuser "username"
* voiceban : Block voice from user "username". Use /voiceban "username" -value. Example /voiceban "rj" -true (could be -false)`

				n = strings.Replace(n, "List of server commands :", "List of server commands : ", -1)

				return n
			}()

			conn, err := rcon.Dial(addr, password)
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}
			defer conn.Close()

			result, err := conn.Execute("help")
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}

			if result != needle {
				t.Fatalf("got result %q, want %q", result, needle)
			}
		})
	}

	if run := getVar("TEST_RUST_SERVER", "false"); run == "true" {
		addr := getVar("TEST_RUST_SERVER_ADDR", "127.0.0.1:28016")
		password := getVar("TEST_RUST_SERVER_PASSWORD", "docker")

		t.Run("rust server", func(t *testing.T) {
			conn, err := rcon.Dial(addr, password)
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}
			defer conn.Close()

			result, err := conn.Execute("status")
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}

			if result == "" {
				t.Fatal("got empty result, want value")
			}

			fmt.Println(result)
		})
	}
}

// getVar returns environment variable or default value.
func getVar(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
