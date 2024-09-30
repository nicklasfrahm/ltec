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

var version = "dev"

const (
	reconciliationInterval = 1 * time.Minute
	requiredArgCount       = 2
)

func main() {
	logger := logging.NewLogger()

	if len(os.Args) < requiredArgCount {
		logger.Error("Missing access point name argument")
		logger.Sugar().Fatalf("Usage: %s <access-point-name>", os.Args[0])
	}

	apn := os.Args[1]
	logger.Info("Starting LTE controller", zap.String("apn", apn), zap.String("version", version))

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
		modemLogger := logger.With(zap.Int64("modem", modem.Index))

		status, err := modem.GetStatus(context.Background())
		if err != nil {
			modemLogger.Error("Failed to get modem status", zap.Error(err))

			continue
		}

		modemLogger.Info("Successfully queried modem status")

		if len(status.Generic.Bearers) == 0 {
			modemLogger.Info("Connecting modem", zap.String("apn", apn))

			if err := modem.SimpleConnect(context.Background(), apn); err != nil {
				modemLogger.Warn("Failed to connect modem", zap.Error(err))

				continue
			}

			// Ensure that the status is updated.
			status, err = modem.GetStatus(context.Background())
			if err != nil {
				modemLogger.Warn("Failed to get modem status", zap.Error(err))

				continue
			}
		}

		for _, bearerDBUSPath := range status.Generic.Bearers {
			if err := reconcileBearer(modemLogger, modem, bearerDBUSPath); err != nil {
				modemLogger.Error("Failed to reconcile bearer", zap.Error(err), zap.String("bearer", bearerDBUSPath))

				continue
			}

			// Implement connectivity test.

			modemLogger.Info("Successfully reconciled modem")

			return nil
		}
	}

	return nil
}

// reconcileBearer ensures that the bearer is connected and working.
func reconcileBearer(logger *zap.Logger, modem *modemmanager.Modem, bearerDBUSPath string) error {
	bearer, err := modem.GetBearer(context.Background(), bearerDBUSPath)
	if err != nil {
		logger.Error("Failed to get bearer status", zap.Error(err), zap.String("bearer", bearerDBUSPath))

		return fmt.Errorf("failed to get bearer status: %w", err)
	}

	logger.Info("Successfully queried bearer status", zap.String("bearer", bearer.DBusPath))

	if !bearer.Status.Connected {
		logger.Info("Connecting bearer", zap.String("bearer", bearer.DBusPath))

		if err := bearer.Connect(context.Background()); err != nil {
			logger.Warn("Failed to connect bearer", zap.String("bearer", bearer.DBusPath), zap.Error(err))

			return fmt.Errorf("failed to connect bearer: %w", err)
		}

		// Ensure that the status is updated.
		bearer, err = modem.GetBearer(context.Background(), bearer.DBusPath)
		if err != nil {
			logger.Warn("Failed to get bearer status", zap.String("bearer", bearer.DBusPath), zap.Error(err))

			return fmt.Errorf("failed to get bearer status: %w", err)
		}
	}

	if err := bearer.ConfigureInterface(context.Background()); err != nil {
		logger.Warn("Failed to configure interface", zap.String("bearer", bearer.DBusPath), zap.Error(err))

		return fmt.Errorf("failed to configure interface: %w", err)
	}

	logger.Info("Successfully configured interface",
		zap.String("bearer", bearer.DBusPath),
		zap.String("interface", bearer.Status.Interface),
	)

	return nil
}
