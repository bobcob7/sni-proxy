package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bobcob7/sni-proxy/internal/config"
	"github.com/bobcob7/sni-proxy/pkg/proxy"
	"go.uber.org/zap"
)

func main() {
	var p proxy.Proxy
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	if err := config.GetConfig(&p); err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return
	}
	ctx, done := context.WithCancel(context.Background())
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-shutdown
		logger.Info("Shutting down")
		done()
	}()
	p.Run(ctx, logger)
}
