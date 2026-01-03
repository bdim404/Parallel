package config

import (
	"fmt"
	"net"
)

type Config struct {
	Listeners []ListenerConfig `json:"listeners"`
}

type ListenerConfig struct {
	Listen string   `json:"listen"`
	Socks  []string `json:"socks"`
}

func (c *Config) Validate() error {
	if len(c.Listeners) == 0 {
		return fmt.Errorf("no listeners configured")
	}

	for i, listener := range c.Listeners {
		if err := listener.Validate(); err != nil {
			return fmt.Errorf("listener %d: %w", i, err)
		}
	}

	return nil
}

func (lc *ListenerConfig) Validate() error {
	if lc.Listen == "" {
		return fmt.Errorf("listen address is empty")
	}

	host, port, err := net.SplitHostPort(lc.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	if host == "" {
		return fmt.Errorf("listen host is empty")
	}

	if port == "" {
		return fmt.Errorf("listen port is empty")
	}

	if len(lc.Socks) == 0 {
		return fmt.Errorf("no socks upstreams configured")
	}

	for i, sock := range lc.Socks {
		_, _, err := net.SplitHostPort(sock)
		if err != nil {
			return fmt.Errorf("invalid socks upstream %d (%s): %w", i, sock, err)
		}
	}

	return nil
}
