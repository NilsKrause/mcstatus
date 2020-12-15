package client

import (
	"time"

	"git.0cd.xyz/michael/mcstatus/pb"
)

// GetStatus gets minecraft server status
func (client *Client) GetStatus() (*pb.Response, error) {
	for {
		if err := client.write(); err != nil {
			return nil, err
		}
		resp, err := client.read()
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

// PingServer pings Minecraft server
func (client *Client) PingServer() time.Duration {
	ping := make([]byte, 1)
	start := time.Now()
	client.Conn.Write([]byte{0x01, 0x00})
	_, _ = client.Conn.Read(ping[:])
	diff := time.Now().Sub(start)
	return diff
}
