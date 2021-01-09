package main

import (
	"klew/creep/handler/socks5"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go socks5.HandleRequest(client)
	}
}
