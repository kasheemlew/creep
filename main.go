package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var hdType string

func init() {
	flag.StringVar(&hdType, "hd", "http", "handler type could be \"http\" or \"socks5\"")
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	flag.Parse()
	var hd func(net.Conn)
	switch hdType {
	case "socks5":
		hd = handleSocket

		println("using socks5")
	case "http":
		fallthrough
	default:
		hd = handleHTTP
		println("using http")
	}
	for {
		client, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go hd(client)
	}
}

func handleHTTP(client net.Conn) {
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

func handleSocket(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()
	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	if b[0] == 0x05 { //只处理Socket5协议
		// 客户端回应：Socket服务端不需要验证方式
		client.Write([]byte{0x05, 0x00})
		n, err = client.Read(b[:])

		// 解析客户端请求访问的目标服务器相关信息
		var host, port string
		switch b[3] { // b[3] 为 ATYP
		case 0x01: // IPv4
			host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		case 0x03: // domain
			host = string(b[5 : n-2])
		case 0x04: // IPv6
			host = net.IP(b[4:19]).String()
		}
		port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
		// 连接远程服务器
		server, err := net.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			panic(err)
		}
		defer server.Close()
		client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
		go io.Copy(server, client)
		io.Copy(client, server)
	}
}
