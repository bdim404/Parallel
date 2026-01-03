package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bdim404/parallel-socks/src/cmd"
	"github.com/bdim404/parallel-socks/src/config"
	"github.com/bdim404/parallel-socks/src/listener"
	"github.com/bdim404/parallel-socks/src/logger"
)

func main() {
	flags, err := cmd.ParseFlags()
	if err != nil {
		logger.Fatal("parse flags: %v", err)
	}

	var cfg *config.Config
	if flags.Config != nil {
		cfg = flags.Config
	} else {
		cfg, err = config.LoadConfig(flags.ConfigPath)
		if err != nil {
			logger.Fatal("load config: %v", err)
		}
	}

	if flags.LogLevel != "" {
		cfg.LogLevel = flags.LogLevel
	}

	if err := cfg.Validate(); err != nil {
		logger.Fatal("validate config: %v", err)
	}

	logger.SetLevel(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	for _, listenerCfg := range cfg.Listeners {
		l, err := listener.New(&listenerCfg)
		if err != nil {
			logger.Fatal("create listener for %s: %v", listenerCfg.Listen, err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Serve(ctx); err != nil {
				logger.Info("listener error: %v", err)
			}
		}()
	}

	<-sigCh
	logger.Info("shutting down...")
	cancel()
	wg.Wait()
	logger.Info("shutdown complete")
}
