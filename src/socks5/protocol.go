package socks5

import (
	"net"
	"strconv"
)

const (
	Version5 = 0x05

	MethodNoAuth = 0x00
	MethodNoAcceptable = 0xFF

	CmdConnect = 0x01

	AtypIPv4 = 0x01
	AtypDomain = 0x03
	AtypIPv6 = 0x04

	RepSuccess = 0x00
	RepGeneralFailure = 0x01
	RepConnectionNotAllowed = 0x02
	RepNetworkUnreachable = 0x03
	RepHostUnreachable = 0x04
	RepConnectionRefused = 0x05
	RepTTLExpired = 0x06
	RepCommandNotSupported = 0x07
	RepAddressTypeNotSupported = 0x08
)

type SOCKS5Error struct {
	ReplyCode byte
	Message   string
}

func (e *SOCKS5Error) Error() string {
	return e.Message
}

type TargetAddress struct {
	Type       byte
	Host       string
	Port       uint16
	RawRequest []byte
}

func (t *TargetAddress) String() string {
	return net.JoinHostPort(t.Host, strconv.Itoa(int(t.Port)))
}

func (t *TargetAddress) Network() string {
	return "tcp"
}
