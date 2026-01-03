package http

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/bdim404/SockRacer/src/config"
	"github.com/bdim404/SockRacer/src/pool"
	"github.com/bdim404/SockRacer/src/socks5"
)

func HandleHTTP(ctx context.Context, reader *bufio.Reader, clientConn net.Conn, target *socks5.TargetAddress, p *pool.Pool) error {
	upstreams := p.GetUpstreams()

	for {
		req, err := ParseRequest(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("parse request failed: %v", err)
			sendErrorResponse(clientConn, 400, "Bad Request")
			return err
		}

		log.Printf("HTTP %s %s", req.Method, req.URI)

		resp, winnerName, err := raceRequest(ctx, req, target, upstreams)
		if err != nil {
			log.Printf("race request failed: %v", err)
			sendErrorResponse(clientConn, 502, "Bad Gateway")
			return err
		}

		log.Printf("✓ %s %s -> %s (%d %s)", req.Method, req.URI, winnerName, resp.StatusCode, resp.Status)

		_, err = clientConn.Write(resp.Raw)
		if err != nil {
			return err
		}

		if req.ShouldClose() || resp.ShouldClose() {
			return nil
		}
	}
}

func raceRequest(ctx context.Context, req *HTTPRequest, target *socks5.TargetAddress, upstreams []config.UpstreamConfig) (*HTTPResponse, string, error) {
	respCh := make(chan *result, len(upstreams))

	raceCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, upstream := range upstreams {
		go func(u config.UpstreamConfig) {
			conn, err := socks5.DialSOCKS5(raceCtx, u.Address, target)
			if err != nil {
				respCh <- &result{err: err, upstream: u}
				return
			}
			defer conn.Close()

			_, err = conn.Write(req.Raw)
			if err != nil {
				respCh <- &result{err: err, upstream: u}
				return
			}

			reader := bufio.NewReader(conn)
			resp, err := ParseResponse(reader)
			if err != nil {
				respCh <- &result{err: err, upstream: u}
				return
			}

			select {
			case respCh <- &result{response: resp, upstream: u}:
			case <-raceCtx.Done():
			}
		}(upstream)
	}

	for i := 0; i < len(upstreams); i++ {
		select {
		case res := <-respCh:
			if res.err == nil && res.response != nil {
				winnerName := res.upstream.Address
				if res.upstream.Name != "" {
					winnerName = fmt.Sprintf("%s (%s)", res.upstream.Name, res.upstream.Address)
				}

				go drainChannel(respCh, len(upstreams)-i-1)
				return res.response, winnerName, nil
			}
		case <-raceCtx.Done():
			return nil, "", fmt.Errorf("request timeout")
		}
	}

	return nil, "", fmt.Errorf("all upstreams failed")
}

type result struct {
	response *HTTPResponse
	upstream config.UpstreamConfig
	err      error
}

func drainChannel(ch chan *result, remaining int) {
	for i := 0; i < remaining; i++ {
		<-ch
	}
}

func sendErrorResponse(conn net.Conn, code int, message string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", code, message)
	conn.Write([]byte(response))
}
