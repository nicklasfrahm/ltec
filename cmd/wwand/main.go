package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nicklasfrahm-dev/appkit/logging"
	"github.com/nicklasfrahm/ltec/pkg/qmi"
	"github.com/nicklasfrahm/ltec/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	logger := logging.NewLogger()

	logger.Info("Starting application", zap.String("version", version))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGPIPE,
	)

	ctx := logging.WithLogger(context.Background(), logger)

	registry := prometheus.NewRegistry()

	components := map[string]service.Component{
		"api_server":     service.NewAPIServer(ctx),
		"metrics_server": service.NewMetricsServer(ctx, registry),
		"modem_manager":  qmi.NewModemManager(ctx, registry),
	}

	for name, component := range components {
		if err := component.Start(ctx); err != nil {
			logger.Fatal("Failed to start component", zap.String("component", name), zap.Error(err))
		}
	}

	// // Create modem instance and start monitoring connection
	// modem := qmi.NewModem("/dev/cdc-wdm0", "apn.example.com", registry)
	// go modem.MonitorConnection()

	// Wait for signals.
	sig := <-signals

	logger.Info("Received signal", zap.String("signal", sig.String()))

	for name, component := range components {
		if err := component.Stop(ctx); err != nil {
			logger.Error("Failed to stop component", zap.String("component", name), zap.Error(err))
		}
	}
}
