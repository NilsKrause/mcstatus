package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cmd := ui()
	conn, err := newConn(cmd)
	if err != nil {
		return err
	}
	for {
		if err := conn.write(); err != nil {
			return err
		}
		resp, err := conn.read()
		if err != nil {
			return err
		}
		if cmd.ping {
			fmt.Printf("Ping: %+v\n", conn.pingServer())
			if err = conn.conn.Close(); err != nil {
				conn.conn.Close()
				return err
			}
			return nil
		}
		if cmd.raw {
			json, err := json.MarshalIndent(&resp, "", "  ")
			if err != nil {
				conn.conn.Close()
				return err
			}
			fmt.Printf("%s\n", string(json))
			conn.conn.Close()
			return err
		}
		fmt.Printf("Name: %s\nPlayers: %d/%d\nVersion: %s\n",
			resp.Description.Text,
			resp.Players.Online,
			resp.Players.Max,
			resp.Version.Name)
		if resp.Players.Online >= 1 {
			fmt.Println("Online:")
			for _, player := range resp.Players.Sample {
				fmt.Printf("\t%s\n", player.Name)
			}
		}
		fmt.Printf("Ping: %+v\n", conn.pingServer())
		if err = conn.conn.Close(); err != nil {
			conn.conn.Close()
			return err
		}
		return nil
	}
}
