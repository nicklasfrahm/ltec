package modemmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

//

// Modem represents a modem.
type Modem struct {
	Index int64
	Path  string
}

// ErrInvalidPath is returned when the path is invalid.
var ErrInvalidPath = errors.New("invalid path")

// NewModem creates a new modem.
func NewModem(path string) (*Modem, error) {
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return nil, fmt.Errorf("failed to create modem: %w: %s", ErrInvalidPath, path)
	}

	index, err := strconv.ParseInt(segments[len(segments)-1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to create modem: %w: %s", err, path)
	}

	modem := &Modem{
		Index: index,
		Path:  path,
	}

	return modem, nil
}

// GetStatus gets the status of the modem.
func (m *Modem) GetStatus(ctx context.Context) (*ModemStatus, error) {
	cmd := exec.CommandContext(ctx, "mmcli", fmt.Sprintf("--modem=%d", m.Index), "--output-json")

	// TODO: Provide more context for errors.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query modem status: %w", err)
	}

	var resp ModemStatusResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode modem status: %w", err)
	}

	return &resp.Modem, nil
}

// SimpleConnect connects the modem to the given APN.
func (m *Modem) SimpleConnect(ctx context.Context, apn string) error {
	cmd := exec.CommandContext(ctx, "mmcli",
		fmt.Sprintf("--modem=%d", m.Index),
		fmt.Sprintf("--simple-connect='apn=%s,ip-type=ipv4v6", apn),
	)

	// TODO: Provide more context for errors.
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to connect modem: %w", err)
	}

	return nil
}

// GetBearer gets the specified bearer of the modem.
func (m *Modem) GetBearer(ctx context.Context, dbusPath string) (*Bearer, error) {
	cmd := exec.CommandContext(ctx, "mmcli", fmt.Sprintf("--modem=%d", m.Index), fmt.Sprintf("--bearer=%s", dbusPath), "--output-json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query bearer: %w", err)
	}

	var resp BearerResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode bearer: %w", err)
	}

	return &resp.Bearer, nil
}
