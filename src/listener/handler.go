package listener

import (
	"context"
	"io"
	"net"

	"github.com/bdim404/parallel-socks/src/logger"
	"github.com/bdim404/parallel-socks/src/pool"
	"github.com/bdim404/parallel-socks/src/socks5"
)

func handleConnection(ctx context.Context, clientConn net.Conn, p *pool.Pool) {
	defer clientConn.Close()

	if err := socks5.HandleNegotiation(clientConn); err != nil {
		logger.Info("negotiation failed: %v", err)
		return
	}

	target, err := socks5.ParseRequest(clientConn)
	if err != nil {
		logger.Info("parse request failed: %v", err)
		socks5.SendReply(clientConn, socks5.RepGeneralFailure, nil)
		return
	}

	if err := socks5.SendReply(clientConn, socks5.RepSuccess, clientConn.LocalAddr()); err != nil {
		logger.Info("send reply failed: %v", err)
		return
	}

	upstreamConn, _, err := p.GetConn(ctx, target)
	if err != nil {
		return
	}
	defer upstreamConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(upstreamConn, clientConn)
		upstreamConn.Close()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(clientConn, upstreamConn)
		clientConn.Close()
		done <- struct{}{}
	}()

	select {
	case <-done:
		<-done
	case <-ctx.Done():
		clientConn.Close()
		upstreamConn.Close()
		<-done
		<-done
	}
}
