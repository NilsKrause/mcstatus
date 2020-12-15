package client

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"git.0cd.xyz/michael/mcstatus/pb"
	"github.com/golang/protobuf/jsonpb"
)

// Client TCP client
type Client struct {
	Addr    string
	Port    int
	Version uint64
	Conn    net.Conn
}

// New client connection
func New(addr string, port int, ver uint64) (*Client, error) {
	conn, err := net.Dial("tcp", addr+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &Client{
		Addr:    addr,
		Port:    port,
		Version: ver,
		Conn:    conn,
	}, nil
}

func (client *Client) write() error {
	if err := client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	client.Conn.Write(handshake(client.Addr, client.Port, client.Version))
	return nil
}

func (client *Client) read() (*pb.Response, error) {
	var response pb.Response
	buf := make([]byte, 1024)
	n, err := client.Conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, err
		}
		return nil, err
	}
	b := bytes.Split(buf[:n], []byte("{"))
	ne := bytes.SplitAfterN(buf[:n], b[0], 2)
	trim := bytes.TrimSuffix(ne[1], []byte("\x00"))
	js := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err := js.Unmarshal(bytes.NewBuffer(trim), &response); err != nil {
		return nil, err
	}
	return &response, nil
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
	handshake.Write([]byte{0x01, 0x00})
	return handshake.Bytes()
}
