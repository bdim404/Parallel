package listener

import (
	"bufio"
	"context"
	"log"
	"net"

	httpracer "github.com/bdim404/SockRacer/src/http"
	"github.com/bdim404/SockRacer/src/pool"
	"github.com/bdim404/SockRacer/src/relay"
	"github.com/bdim404/SockRacer/src/socks5"
)

func handleConnection(ctx context.Context, clientConn net.Conn, p *pool.Pool) {
	defer clientConn.Close()

	if err := socks5.HandleNegotiation(clientConn); err != nil {
		log.Printf("negotiation failed: %v", err)
		return
	}

	target, err := socks5.ParseRequest(clientConn)
	if err != nil {
		log.Printf("parse request failed: %v", err)
		socks5.SendReply(clientConn, socks5.RepGeneralFailure, nil)
		return
	}

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, clientConn.LocalAddr()); err != nil {
		log.Printf("send reply failed: %v", err)
		return
	}

	reader := bufio.NewReader(clientConn)
	isHTTP, err := httpracer.DetectProtocol(reader)
	if err != nil {
		log.Printf("protocol detection failed: %v", err)
		return
	}

	if isHTTP {
		log.Printf("HTTP mode for %s", target)
		if err := httpracer.HandleHTTP(ctx, reader, clientConn, target, p); err != nil {
			log.Printf("HTTP racing failed: %v", err)
		}
	} else {
		log.Printf("HTTPS/relay mode for %s", target)

		upstreamConn, err := p.GetConn(ctx, target)
		if err != nil {
			if socks5Err, ok := err.(*socks5.SOCKS5Error); ok {
				log.Printf("✗ %s upstream failed: SOCKS5 reply code %d", target, socks5Err.ReplyCode)
			} else {
				log.Printf("✗ %s upstream failed: %v", target, err)
			}
			return
		}
		defer upstreamConn.Close()

		relay.BidirectionalWithReader(ctx, reader, clientConn, upstreamConn)
	}
}
