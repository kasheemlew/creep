/*
* SOCKS Protocol Version 5: https://tools.ietf.org/html/rfc1928
 */

package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
)

func HandleRequest(client net.Conn) {

	remoteAddr := client.RemoteAddr().String()
	logrus.Info("Remote addr: ", remoteAddr)

	if err := Auth(client); err != nil {
		client.Write([]byte(err.Error()))
		client.Close()
		return
	}

}

func Auth(client net.Conn) error {
	/* client connect and sends:
	 *
	 * +----+----------+----------+
	 * |VER | NMETHODS | METHODS  |
	 * +----+----------+----------+
	 * | 1  |    1     | 1 to 255 |
	 * +----+----------+----------+
	 */

	buf := make([]byte, 256)

	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	// ensure socks version is 5
	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// read methods
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	/* server reply NOAUTH:
	 *
	 * +----+--------+
	 * |VER | METHOD |
	 * +----+--------+
	 * | 1  |   1    |
	 * +----+--------+ */
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp err: " + err.Error())
	}

	return nil
}

func Connect(client net.Conn) (net.Conn, error) {
	/* client sends connect request:
	 *
	 * +----+-----+-------+------+----------+----------+
	 * |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	 * +----+-----+-------+------+----------+----------+
	 * | 1  |  1  | X'00' |  1   | Variable |    2     |
	 * +----+-----+-------+------+----------+----------+ */
	var buf [256]byte

	n, err := io.ReadFull(client, buf[:4])
	if n != 4 {
		return nil, errors.New("socks connect read head error: " + err.Error())
	}
	// RSV could be ignored
	ver, cmd, atyp := int(buf[0]), int(buf[1]), int(buf[3])
	if ver != 5 {
		return nil, errors.New("socks connect version not supported: " + strconv.Itoa(ver))
	}
	// TODO: cmd should support BIND(2) and UDP ASSOCIATE(3)
	if cmd != 1 {
		return nil, errors.New("socks connect only support cmd: CONNECT, not supported: " + strconv.Itoa(cmd))
	}

	/* ATYP:
	 * 1: IPv4 with a length of 4 octets
	 * 3: DOMAIN, first octet contains the number of octets of name that follow
	 * 4: IPv6 with a length of 16 octets
	 */
	var addr string
	switch atyp {
	case 1:
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 {
			return nil, errors.New("invalid IPv4: " + err.Error())
		}
		addr = net.IPv4(buf[0], buf[1], buf[2], buf[3]).String()
	case 3:
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])
		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])
	case 4:
		return nil, errors.New("IPv6 not supported")
	}

	// read port from last 2 octets
	n, err = io.ReadFull(client, buf[:2])
	if n != 2 {
		return nil, errors.New("invalid port: " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	// connect remote
	dest, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, errors.New("dial failed: " + err.Error())
	}

	/* Reply to client
	 *
	 * +----+-----+-------+------+----------+----------+
	 * |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	 * +----+-----+-------+------+----------+----------+
	 * | 1  |  1  | X'00' |  1   | Variable |    2     |
	 * +----+-----+-------+------+----------+----------+ */

	// client will work even addr and port are all filled with 0
	n, err = client.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	if err != nil {
		dest.Close()
		return nil, errors.New("write resp failed: " + err.Error())
	}
	return dest, nil
}

func Forward(client, dest net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, dest)
	go forward(dest, client)
}
