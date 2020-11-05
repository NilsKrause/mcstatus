package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
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

func pingServer(conn net.Conn) time.Duration {
	ping := make([]byte, 1)
	start := time.Now()
	conn.Write([]byte{0x01, 0x00})
	_, _ = conn.Read(ping[:])
	diff := time.Now().Sub(start)
	return diff
}

func orig() error {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	addr := flag.String("addr", "127.0.0.1", "Server address")
	port := flag.Int("port", 25565, "Server Port")
	ver := flag.Uint64("ver", 751, "Minecraft protocol version number")
	raw := flag.Bool("raw", false, "Prints raw json")
	ping := flag.Bool("ping", false, "Pings the server")
	flag.Parse()

	/*if len(os.Args) < 2 {
		flag.Usage()
		return
	}*/

	conn, err := net.Dial("tcp", *addr+":"+strconv.Itoa(*port))
	if err != nil {
		return err
	}

	for {
		err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			conn.Close()
			return err
		}

		conn.Write(handshake(*addr, *port, *ver))
		conn.Write([]byte{0x01, 0x00})

		err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			conn.Close()
			return err
		}

		recvBuf := make([]byte, 512)
		var resp response

		n, err := conn.Read(recvBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				conn.Close()
				return err
			}
			conn.Close()
			return err
		}

		b := bytes.Split(recvBuf[:n], []byte("{"))
		ne := bytes.SplitAfterN(recvBuf[:n], b[0], 2)
		trim := bytes.TrimSuffix(ne[1], []byte("\x00"))

		if *ping == false {
			if err := json.Unmarshal(trim, &resp); err != nil {
				conn.Close()
				return err
			}
			if *raw == true {
				json, err := json.MarshalIndent(resp, "", "  ")
				if err != nil {
					conn.Close()
					return err
				}
				fmt.Printf("%s\n", string(json))
				conn.Close()
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
		}
		fmt.Printf("Ping: %+v\n", pingServer(conn))
		if err = conn.Close(); err != nil {
			conn.Close()
			return err
		}
		return nil
	}
}

func main() {
	if err := orig(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	addr := flag.String("addr", "127.0.0.1", "Server address")
	port := flag.Int("port", 25565, "Server Port")
	//ver := flag.Uint64("ver", 751, "Minecraft protocol version number")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr+":"+strconv.Itoa(*port))
	if err != nil {
		return err
	}

	/*conn.Write(handshake(*addr, *port, *ver))
	conn.Write([]byte{0x01, 0x00})*/

	recvBuf := make([]byte, 512)

	n, err := conn.Read(recvBuf)
	if err != nil {
		return err
	}

	conn.Write([]byte{0x00})

	fmt.Println(string(recvBuf[:n]))
	return nil
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
