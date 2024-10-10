package modemmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nicklasfrahm-dev/appkit/logging"
	"go.uber.org/zap"
)

const defaultTimeout = 100 * time.Millisecond

// ErrNoModems is returned when no modems are found.
var ErrNoModems = errors.New("no modems detected")

// ModemListResponse represents a modem list response.
type ModemListResponse struct {
	// ModemList is a list of modem paths.
	ModemList []string `json:"modem-list"`
}

// ListModems lists all modems.
func ListModems(ctx context.Context) ([]*Modem, error) {
	timeoutContext, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutContext, "mmcli", "--list-modems", "--output-json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		logging.FromContext(ctx).Error("Failed to run command", zap.String("output", strings.TrimSpace(string(output))))

		return nil, fmt.Errorf("failed to list modems: %w", err)
	}

	var resp ModemListResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode modem list: %w", err)
	}

	if len(resp.ModemList) == 0 {
		return nil, ErrNoModems
	}

	modems := make([]*Modem, 0, len(resp.ModemList))

	for _, path := range resp.ModemList {
		modem, err := NewModem(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create modem: %w", err)
		}

		modems = append(modems, modem)
	}

	return modems, nil
}
