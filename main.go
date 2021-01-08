package main

import (
	"encoding/json"
	"fmt"
	"os"

	"git.0cd.xyz/michael/mcstatus/client"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cmd := ui()
	client, err := client.New(cmd.addr, cmd.port, cmd.version)
	if err != nil {
		return err
	}
	defer client.Conn.Close()
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
			json, err := json.MarshalIndent(&status, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(json))
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
		return nil
	}
}
