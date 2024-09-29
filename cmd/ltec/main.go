package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nicklasfrahm-dev/appkit/logging"
	"github.com/nicklasfrahm/ltec/pkg/modemmanager"
	"go.uber.org/zap"
)

const reconciliationInterval = 1 * time.Minute

func main() {
	logger := logging.NewLogger()

	if len(os.Args) < 2 {
		logger.Error("Missing access point name argument")
		logger.Sugar().Fatalf("Usage: %s <access-point-name>", os.Args[0])
	}

	apn := os.Args[1]
	logger.Info("Starting LTE controller", zap.String("apn", apn))

	if err := reconcile(logger, apn); err != nil {
		logger.Error("Failed to reconcile", zap.Error(err), zap.String("trigger", "startup"))
	}

	for tick := range time.Tick(reconciliationInterval) {
		if err := reconcile(logger, apn); err != nil {
			logger.Error("Failed to reconcile", zap.Error(err), zap.String("trigger", "interval"), zap.Time("tick", tick))
		}
	}
}

func reconcile(logger *zap.Logger, apn string) error {
	modems, err := modemmanager.ListModems(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list modems: %w", err)
	}

	for _, modem := range modems {
		status, err := modem.GetStatus(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get modem status: %w", err)
		}

		logger.Info("Successfully queried modem status", zap.Int64("modem", modem.Index))

		if len(status.Generic.Bearers) == 0 {
			logger.Info("Connecting modem", zap.String("apn", apn))
			// TODO: Connect to modem.

			// TODO: Update modem status.
			status, err = modem.GetStatus(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get modem status: %w", err)
			}
		}

		logger.Info("Modem is connected", zap.Int64("modem", modem.Index), zap.String("state", status.Generic.State))

	}

	return nil
}
