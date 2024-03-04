package rcontest_test

import (
	"fmt"
	"log"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
)

func ExampleServer() {
	server := rcontest.NewServer(
		rcontest.SetSettings(rcontest.Settings{Password: "password"}),
		rcontest.SetCommandHandler(func(c *rcontest.Context) {
			switch c.Request().Body() {
			case "Hello, server":
				rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "Hello, client").WriteTo(c.Conn())
			default:
				rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, "unknown command").WriteTo(c.Conn())
			}
		}),
	)
	defer server.Close()

	client, err := rcon.Dial(server.Addr(), "password")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	response, err := client.Execute("Hello, server")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)

	response, err = client.Execute("Hi!")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)

	// Output:
	// Hello, client
	// unknown command
}
