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
	"go.uber.org/zap"
)

// DefaultPort is the default port of the API component.
const DefaultPort = 8080

// APIServer represents an APIServer component.
type APIServer struct {
	app    *echo.Echo
	logger *zap.Logger
}

// Start starts the API component.
func (api *APIServer) Start(ctx context.Context) error {
	port := api.GetPort()

	api.logger.Info("Starting component", zap.Int("port", port))

	go func() {
		if err := api.app.Start(":" + strconv.Itoa(port)); err != nil {
			// This is expected behavior when we shut down the server.
			if errors.Is(err, http.ErrServerClosed) {
				return
			}

			api.logger.Error("Failed to start component", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the API component.
func (api *APIServer) Stop(ctx context.Context) error {
	api.logger.Info("Stopping component")

	if err := api.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop component: %w", err)
	}

	return nil
}

// NewAPIServer creates a new API component.
func NewAPIServer(ctx context.Context) *APIServer {
	app := echo.New()
	app.HideBanner = true
	app.HidePort = true

	app.Any("/*", NewHealthCheckHandler())

	return &APIServer{
		app:    app,
		logger: logging.FromContext(ctx).With(zap.String("component", "api_server")),
	}
}

// GetPort returns the port of the API component.
func (api *APIServer) GetPort() int {
	rawPort := os.Getenv("PORT")
	if rawPort == "" {
		return DefaultPort
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil {
		api.logger.Warn("Failed to parse port", zap.String("raw_port", rawPort), zap.Error(err))

		api.logger.Warn("Using default port", zap.Int("port", DefaultPort))

		return DefaultPort
	}

	return port
}
