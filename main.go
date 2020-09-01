package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

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

func main() {
	addr := flag.String("addr", "127.0.0.1", "Server address")
	port := flag.Int("port", 25565, "Server Port")
	ver := flag.Uint64("ver", 751, "Minecraft protocol version number")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		err := conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			log.Println("WriteDealine failed:", err)
			return
		}

		conn.Write(handshake(*addr, *port, *ver))
		conn.Write([]byte{0x01, 0x00})

		recvBuf := make([]byte, 512)
		var resp response

		err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			log.Println("SetReadDealine failed:", err)
			return
		}

		n, err := conn.Read(recvBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("read timeout:", err)
			} else {
				log.Println("read error", err)
			}
		}

		requestright := recvBuf[:n]
		b := bytes.Split(requestright, []byte("{"))
		ne := bytes.SplitAfterN(requestright, b[0], 2)
		after := bytes.TrimSuffix(ne[1], []byte("\x00"))

		if err := json.Unmarshal(after, &resp); err != nil {
			log.Println(err)
			conn.Close()
			return
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

		recvBuf = make([]byte, 1)
		start := time.Now()
		conn.Write([]byte{0x01, 0x00})
		_, _ = conn.Read(recvBuf[:])
		diff := time.Now().Sub(start)
		fmt.Printf("Ping: %+v\n", diff)
		if err = conn.Close(); err != nil {
			log.Println(err)
			return
		}
		return
	}
}

type response struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon string `json:"favicon"`
}
