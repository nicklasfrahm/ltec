package service

import (
	"context"
)

// Component represents a service component.
type Component interface {
	// Start starts the component.
	Start(ctx context.Context) error
	// Stop stops the component.
	Stop(ctx context.Context) error
}
