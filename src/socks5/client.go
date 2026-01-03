package socks5

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func DialSOCKS5(ctx context.Context, proxyAddr string, target *TargetAddress) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("dial proxy: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	if err := clientNegotiate(conn); err != nil {
		conn.Close()
		return nil, err
	}

	if err := clientConnect(conn, target); err != nil {
		conn.Close()
		return nil, err
	}

	conn.SetDeadline(time.Time{})
	return conn, nil
}

func clientNegotiate(conn net.Conn) error {
	_, err := conn.Write([]byte{Version5, 1, MethodNoAuth})
	if err != nil {
		return fmt.Errorf("write negotiation: %w", err)
	}

	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("read negotiation response: %w", err)
	}

	if buf[0] != Version5 {
		return fmt.Errorf("unsupported version: %d", buf[0])
	}

	if buf[1] != MethodNoAuth {
		return fmt.Errorf("unsupported auth method: %d", buf[1])
	}

	return nil
}

func clientConnect(conn net.Conn, target *TargetAddress) error {
	req := make([]byte, 0, 262)
	req = append(req, Version5, CmdConnect, 0x00)

	switch target.Type {
	case AtypIPv4:
		req = append(req, AtypIPv4)
		ip := net.ParseIP(target.Host)
		if ip == nil {
			return fmt.Errorf("invalid IPv4 address: %s", target.Host)
		}
		req = append(req, ip.To4()...)

	case AtypIPv6:
		req = append(req, AtypIPv6)
		ip := net.ParseIP(target.Host)
		if ip == nil {
			return fmt.Errorf("invalid IPv6 address: %s", target.Host)
		}
		req = append(req, ip.To16()...)

	case AtypDomain:
		req = append(req, AtypDomain)
		if len(target.Host) > 255 {
			return fmt.Errorf("domain name too long: %d", len(target.Host))
		}
		req = append(req, byte(len(target.Host)))
		req = append(req, []byte(target.Host)...)

	default:
		return fmt.Errorf("unsupported address type: %d", target.Type)
	}

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, target.Port)
	req = append(req, portBuf...)

	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("write connect request: %w", err)
	}

	reply := make([]byte, 4)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return fmt.Errorf("read reply header: %w", err)
	}

	if reply[0] != Version5 {
		return fmt.Errorf("unsupported version: %d", reply[0])
	}

	if reply[1] != RepSuccess {
		return fmt.Errorf("connection failed: reply code %d", reply[1])
	}

	atyp := reply[3]
	switch atyp {
	case AtypIPv4:
		discard := make([]byte, 6)
		io.ReadFull(conn, discard)
	case AtypIPv6:
		discard := make([]byte, 18)
		io.ReadFull(conn, discard)
	case AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return fmt.Errorf("read bind domain length: %w", err)
		}
		discard := make([]byte, int(lenBuf[0])+2)
		io.ReadFull(conn, discard)
	}

	return nil
}
