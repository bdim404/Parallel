package cmd

import (
	"flag"
	"fmt"
	"net"

	"github.com/bdim404/parallel/src/config"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func ParseFlags() (*config.Config, error) {
	var listenAddr string
	var listenPort string
	var socks stringSlice

	flag.StringVar(&listenAddr, "listen-address", "127.0.0.1", "Listen address")
	flag.StringVar(&listenPort, "listen-port", "", "Listen port")
	flag.Var(&socks, "socks", "Upstream SOCKS5 proxy (can be specified multiple times)")
	flag.Parse()

	if listenPort == "" {
		return nil, nil
	}

	if len(socks) == 0 {
		return nil, fmt.Errorf("at least one --socks upstream must be specified")
	}

	listen := net.JoinHostPort(listenAddr, listenPort)

	cfg := &config.Config{
		Listeners: []config.ListenerConfig{
			{
				Listen: listen,
				Socks:  socks,
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
