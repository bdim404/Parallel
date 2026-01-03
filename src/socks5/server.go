package socks5

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func HandleNegotiation(conn net.Conn) error {
	buf := make([]byte, 257)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return fmt.Errorf("read version and nmethods: %w", err)
	}

	version := buf[0]
	nmethods := buf[1]

	if version != Version5 {
		return fmt.Errorf("unsupported version: %d", version)
	}

	if nmethods == 0 {
		return fmt.Errorf("no methods provided")
	}

	if _, err := io.ReadFull(conn, buf[:nmethods]); err != nil {
		return fmt.Errorf("read methods: %w", err)
	}

	hasNoAuth := false
	for i := byte(0); i < nmethods; i++ {
		if buf[i] == MethodNoAuth {
			hasNoAuth = true
			break
		}
	}

	if !hasNoAuth {
		conn.Write([]byte{Version5, MethodNoAcceptable})
		return fmt.Errorf("no acceptable methods")
	}

	_, err := conn.Write([]byte{Version5, MethodNoAuth})
	return err
}

func ParseRequest(conn net.Conn) (*TargetAddress, error) {
	var rawBuffer bytes.Buffer

	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, fmt.Errorf("read request header: %w", err)
	}

	version := buf[0]
	cmd := buf[1]
	atyp := buf[3]

	rawBuffer.WriteByte(atyp)

	if version != Version5 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	if cmd != CmdConnect {
		return nil, fmt.Errorf("unsupported command: %d", cmd)
	}

	var host string
	switch atyp {
	case AtypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("read IPv4 address: %w", err)
		}
		rawBuffer.Write(addr)
		host = net.IP(addr).String()

	case AtypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("read IPv6 address: %w", err)
		}
		rawBuffer.Write(addr)
		host = net.IP(addr).String()

	case AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, fmt.Errorf("read domain length: %w", err)
		}
		rawBuffer.Write(lenBuf)
		domainLen := lenBuf[0]
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("read domain: %w", err)
		}
		rawBuffer.Write(domain)
		host = string(domain)

	default:
		return nil, fmt.Errorf("unsupported address type: %d", atyp)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, fmt.Errorf("read port: %w", err)
	}
	rawBuffer.Write(portBuf)
	port := binary.BigEndian.Uint16(portBuf)

	return &TargetAddress{
		Type:       atyp,
		Host:       host,
		Port:       port,
		RawRequest: rawBuffer.Bytes(),
	}, nil
}

func SendReply(conn net.Conn, status byte, bindAddr net.Addr) error {
	reply := make([]byte, 0, 22)
	reply = append(reply, Version5, status, 0x00)

	if bindAddr == nil {
		reply = append(reply, AtypIPv4)
		reply = append(reply, 0, 0, 0, 0)
		reply = append(reply, 0, 0)
	} else {
		tcpAddr, ok := bindAddr.(*net.TCPAddr)
		if !ok {
			reply = append(reply, AtypIPv4)
			reply = append(reply, 0, 0, 0, 0)
			reply = append(reply, 0, 0)
		} else {
			ip4 := tcpAddr.IP.To4()
			if ip4 != nil {
				reply = append(reply, AtypIPv4)
				reply = append(reply, ip4...)
			} else {
				reply = append(reply, AtypIPv6)
				reply = append(reply, tcpAddr.IP.To16()...)
			}
			portBuf := make([]byte, 2)
			binary.BigEndian.PutUint16(portBuf, uint16(tcpAddr.Port))
			reply = append(reply, portBuf...)
		}
	}

	_, err := conn.Write(reply)
	return err
}
