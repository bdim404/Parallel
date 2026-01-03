package racer

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/bdim404/parallel/src/socks5"
)

type Racer struct {
	upstreams []string
	timeout   time.Duration
}

func New(upstreams []string, timeout time.Duration) *Racer {
	return &Racer{
		upstreams: upstreams,
		timeout:   timeout,
	}
}

type raceResult struct {
	conn      net.Conn
	proxyAddr string
	err       error
	duration  time.Duration
}

func (r *Racer) Race(ctx context.Context, target *socks5.TargetAddress) (net.Conn, error) {
	raceCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	resultCh := make(chan *raceResult, len(r.upstreams))
	startTime := time.Now()

	log.Printf("racing %d upstreams for %s", len(r.upstreams), target)

	for _, upstream := range r.upstreams {
		go func(proxy string) {
			connStart := time.Now()
			conn, err := socks5.DialSOCKS5(raceCtx, proxy, target)
			duration := time.Since(connStart)
			resultCh <- &raceResult{
				conn:      conn,
				proxyAddr: proxy,
				err:       err,
				duration:  duration,
			}
		}(upstream)
	}

	var firstConn net.Conn
	var winnerProxy string
	var winnerDuration time.Duration
	var errors []string
	receivedCount := 0

	for i := 0; i < len(r.upstreams); i++ {
		result := <-resultCh
		receivedCount++

		if result.err == nil {
			if firstConn == nil {
				firstConn = result.conn
				winnerProxy = result.proxyAddr
				winnerDuration = result.duration
				log.Printf("✓ winner: %s (%dms) - received %d/%d responses",
					result.proxyAddr, result.duration.Milliseconds(), receivedCount, len(r.upstreams))
			} else {
				log.Printf("  closed: %s (%dms) - slower than winner",
					result.proxyAddr, result.duration.Milliseconds())
				result.conn.Close()
			}
		} else {
			log.Printf("✗ failed: %s (%dms) - %v",
				result.proxyAddr, result.duration.Milliseconds(), result.err)
			errors = append(errors, fmt.Sprintf("%s: %v", result.proxyAddr, result.err))
		}
	}

	totalDuration := time.Since(startTime)

	if firstConn != nil {
		log.Printf("race completed for %s: winner=%s, duration=%dms, total=%dms",
			target, winnerProxy, winnerDuration.Milliseconds(), totalDuration.Milliseconds())
		return firstConn, nil
	}

	return nil, fmt.Errorf("all upstreams failed: %s", strings.Join(errors, "; "))
}
