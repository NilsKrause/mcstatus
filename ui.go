package main

import (
	"flag"
	"fmt"
	"os"
)

type cmd struct {
	addr    	string
	port    	int
	version 	uint64
	raw     	bool
	ping    	bool
	server  	bool
	serverport 	int
}

func ui() *cmd {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	addr := flag.String("addr", "127.0.0.1", "Server address")
	port := flag.Int("port", 25565, "Server Port")
	version := flag.Uint64("ver", 751, "Minecraft protocol version number")
	raw := flag.Bool("raw", false, "Prints raw json")
	ping := flag.Bool("ping", false, "Pings the server")
	server := flag.Bool("server", false, "Serves /status {address, port, version} endpoint (all other args, except serverport) are ignored")
	serverport := flag.Int("serverport", 25566, "Server Port when server flag is set")
	flag.Parse()
	return &cmd{
		addr:    		*addr,
		port:    		*port,
		version: 		*version,
		raw:     		*raw,
		ping:    		*ping,
		server:  		*server,
		serverport: 	*serverport,
	}
}
