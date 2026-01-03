package relay

import (
	"context"
	"io"
	"net"
)

func Bidirectional(ctx context.Context, a, b net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(a, b)
		a.Close()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(b, a)
		b.Close()
		done <- struct{}{}
	}()

	select {
	case <-done:
		<-done
	case <-ctx.Done():
		a.Close()
		b.Close()
		<-done
		<-done
	}
}
