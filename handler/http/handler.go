package http

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

func HandleRequest(client net.Conn) {
	// 读取请求报文
	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return

	}

	// 读取报文中 method, host
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" {
		address = hostPortURL.Scheme + ":443"
	} else {
		address = hostPortURL.Host
		if !strings.Contains(address, ":") {
			address += ":80"
		}
	}

	// 与目标服务器建立连接
	server, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		server.Write(b[:n])
	}
	go io.Copy(server, client)
	io.Copy(client, server)
}
