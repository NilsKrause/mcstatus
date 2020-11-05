package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"git.0cd.xyz/michael/mcstatus/pb"
	"github.com/golang/protobuf/jsonpb"
)

type client struct {
	cmd  *cmd
	conn net.Conn
}

func newConn(cmd *cmd) (*client, error) {
	conn, err := net.Dial("tcp", cmd.addr+":"+strconv.Itoa(cmd.port))
	if err != nil {
		return nil, err
	}
	return &client{
		cmd:  cmd,
		conn: conn,
	}, nil
}

func (client *client) write() error {
	if err := client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		client.conn.Close()
		return err
	}
	client.conn.Write(handshake(client.cmd.addr, client.cmd.port, client.cmd.version))
	client.conn.Write([]byte{0x01, 0x00})
	return nil
}

func (client *client) read() (*pb.Response, error) {
	var response pb.Response
	if err := client.conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		client.conn.Close()
		return nil, err
	}
	recvBuf := make([]byte, 512)
	n, err := client.conn.Read(recvBuf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			client.conn.Close()
			return nil, err
		}
		client.conn.Close()
		return nil, err
	}
	b := bytes.Split(recvBuf[:n], []byte("{"))
	ne := bytes.SplitAfterN(recvBuf[:n], b[0], 2)
	trim := bytes.TrimSuffix(ne[1], []byte("\x00"))
	js := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err := js.Unmarshal(bytes.NewBuffer(trim), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *client) pingServer() time.Duration {
	ping := make([]byte, 1)
	start := time.Now()
	client.conn.Write([]byte{0x01, 0x00})
	_, _ = client.conn.Read(ping[:])
	diff := time.Now().Sub(start)
	return diff
}

func packetlength(b ...[]byte) (length int) {
	for _, bytes := range b {
		length += len(bytes)
	}
	return length
}

func handshake(addr string, port int, ver uint64) []byte {
	id := []byte{0x00}
	state := []byte{0x01}

	version := make([]byte, 2)
	binary.PutUvarint(version, ver)
	p := make([]byte, 2)
	binary.BigEndian.PutUint16(p, uint16(port))
	length := packetlength(id, version, []byte(addr), []byte(p), state) + 1

	var handshake bytes.Buffer
	handshake.WriteByte(byte(length))
	handshake.Write(id)
	handshake.Write(version)
	handshake.WriteByte(byte(len(addr)))
	handshake.WriteString(addr)
	handshake.Write(p)
	handshake.Write(state)
	return handshake.Bytes()
}
