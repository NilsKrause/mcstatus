package client

import (
	"fmt"
	"time"

	"git.0cd.xyz/michael/mcstatus/mcstatuspb"
)

// GetStatus gets minecraft server status
func (client *Client) GetStatus() (*mcstatuspb.Response, error) {
	for {
		if err := client.write(); err != nil {
			fmt.Println("write error")
			return nil, err
		}
		resp, err := client.read()
		if err != nil {
			fmt.Println("read error")
			return nil, err
		}
		return resp, nil
	}
}

// PingServer pings Minecraft server
func (client *Client) PingServer() (time.Duration, error) {
	ping := make([]byte, 1)
	start := time.Now()
	_, err := client.Conn.Write([]byte{0x01, 0x00})
	if err != nil {
		return 0, err
	}
	_, _ = client.Conn.Read(ping[:])
	diff := time.Now().Sub(start)
	return diff, nil
}
