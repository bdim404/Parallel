package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bdim404/parallel/src/cmd"
	"github.com/bdim404/parallel/src/config"
	"github.com/bdim404/parallel/src/listener"
)

func main() {
	cfg, err := cmd.ParseFlags()
	if err != nil {
		log.Fatalf("parse flags: %v", err)
	}

	if cfg == nil {
		cfg, err = config.LoadConfig("config.json")
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	for _, listenerCfg := range cfg.Listeners {
		l, err := listener.New(&listenerCfg)
		if err != nil {
			log.Fatalf("create listener for %s: %v", listenerCfg.Listen, err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Serve(ctx); err != nil {
				log.Printf("listener error: %v", err)
			}
		}()
	}

	<-sigCh
	log.Println("shutting down...")
	cancel()
	wg.Wait()
	log.Println("shutdown complete")
}
