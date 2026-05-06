package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/cronwatch/internal/alert"
	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/monitor"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	alerter := alert.NewManager()
	alerter.Register(alert.NewLogNotifier())

	if cfg.Webhook.URL != "" {
		alerter.Register(alert.NewWebhookNotifier(cfg.Webhook.URL, cfg.Webhook.Secret))
	}

	m, err := monitor.New(cfg, alerter)
	if err != nil {
		log.Fatalf("failed to create monitor: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("cronwatch started, monitoring %d jobs", len(cfg.Jobs))
	m.Run(ctx)
	log.Println("cronwatch stopped")
}
