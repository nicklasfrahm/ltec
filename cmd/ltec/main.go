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

		modemLogger := logger.With(zap.Int64("modem", modem.Index))

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
			bearer, err := modem.GetBearer(context.Background(), bearerDBUSPath)
			if err != nil {
				modemLogger.Warn("Failed to get bearer status", zap.String("bearer", bearerDBUSPath), zap.Error(err))

				continue
			}

			modemLogger.Info("Successfully queried bearer status", zap.String("bearer", bearer.DBusPath))

			if !bearer.Status.Connected {
				modemLogger.Info("Connecting bearer", zap.String("bearer", bearer.DBusPath))

				if err := bearer.Connect(context.Background()); err != nil {
					modemLogger.Warn("Failed to connect bearer", zap.String("bearer", bearer.DBusPath), zap.Error(err))

					continue
				}

				// Ensure that the status is updated.
				bearer, err = modem.GetBearer(context.Background(), bearer.DBusPath)
				if err != nil {
					modemLogger.Warn("Failed to get bearer status", zap.String("bearer", bearer.DBusPath), zap.Error(err))

					continue
				}
			}

			if err := bearer.ConfigureInterface(context.Background()); err != nil {
				modemLogger.Warn("Failed to configure interface", zap.String("bearer", bearer.DBusPath), zap.Error(err))

				continue
			}

			modemLogger.Info("Successfully configured interface", zap.String("bearer", bearer.DBusPath), zap.String("interface", bearer.Status.Interface))
		}

		// TODO: Test connectivity.

		modemLogger.Info("Successfully reconciled modem")

		return nil
	}

	return nil
}
