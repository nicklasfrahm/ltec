package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nicklasfrahm-dev/appkit/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const DefaultMetricsPort = 9000

// MetricsServer represents a MetricsServer component.
type MetricsServer struct {
	app    *echo.Echo
	logger *zap.Logger
}

// Start starts the MetricsServer component.
func (m *MetricsServer) Start(ctx context.Context) error {
	port := m.GetPort()

	m.logger.Info("Starting component", zap.Int("port", port))

	go func() {
		if err := m.app.Start(":" + strconv.Itoa(port)); err != nil {
			// This is expected behavior when we shut down the server.
			if errors.Is(err, http.ErrServerClosed) {
				return
			}

			m.logger.Error("Failed to start component", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the MetricsServer component.
func (m *MetricsServer) Stop(ctx context.Context) error {
	m.logger.Info("Stopping component")

	if err := m.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop component: %w", err)
	}

	return nil
}

// NewMetricsServer creates a new MetricsServer component.
func NewMetricsServer(ctx context.Context, registry *prometheus.Registry) *MetricsServer {
	app := echo.New()
	app.HideBanner = true
	app.HidePort = true

	app.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	app.Any("/*", NewCatchAllHandler())

	return &MetricsServer{
		app:    app,
		logger: logging.FromContext(ctx).With(zap.String("component", "metrics_server")),
	}
}

// GetPort gets the port of the MetricsServer component.
func (m *MetricsServer) GetPort() int {
	rawPort := os.Getenv("METRICS_PORT")
	if rawPort == "" {
		return DefaultMetricsPort
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil {
		m.logger.Warn("Failed to parse port", zap.String("raw_port", rawPort), zap.Error(err))

		m.logger.Warn("Using default port", zap.Int("port", DefaultMetricsPort))

		return DefaultMetricsPort
	}

	return port
}
