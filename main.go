package main

import (
	"encoding/json"
	"fmt"
	"git.0cd.xyz/michael/mcstatus/api"
	"net"
	"os"
	"os/signal"

	mcclient "git.0cd.xyz/michael/mcstatus/client"
)

func main() {
	cmd := ui()

	if cmd.server {
		runServer(cmd)
		return
	}

	if err := runOnce(cmd); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cmd) {
	a := api.NewApi()

	a.Start(cmd.port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	a.Stop()
}

func runOnce(cmd *cmd) error {
	client, err := mcclient.New(cmd.addr, cmd.port, cmd.version)
	if err != nil {
		return err
	}

	defer func(Conn net.Conn) {
		_ = Conn.Close()
	}(client.Conn)

	for {
		status, err := client.GetStatus()
		if err != nil {
			return err
		}
		if cmd.ping {
			ping, err := client.PingServer()
			if err != nil {
				return err
			}
			fmt.Printf("Ping: %+v\n", ping)
			return nil
		}
		if cmd.raw {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(status); err != nil {
				return err
			}
			return nil
		}
		name := status.Description.Text
		if name == "" {
			name = status.Description.Extra[0].Text
		}
		fmt.Printf("Name: %s\nPlayers: %d/%d\nVersion: %s\n",
			name,
			status.Players.Online,
			status.Players.Max,
			status.Version.Name)
		if status.Players.Online >= 1 {
			fmt.Println("Online:")
			for _, player := range status.Players.Sample {
				fmt.Printf("\t%s\n", player.Name)
			}
		}
		ping, err := client.PingServer()
		if err != nil {
			return err
		}
		fmt.Printf("Ping: %+v\n", ping)
	}
}
