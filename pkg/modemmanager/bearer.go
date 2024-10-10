package modemmanager

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/nicklasfrahm-dev/appkit/logging"
	"github.com/nicklasfrahm/ltec/pkg/ip"
	"go.uber.org/zap"
)

// BearerResponse represents a bearer response.
type BearerResponse struct {
	// Bearer contains the bearer.
	Bearer Bearer `json:"bearer"`
}

// BearerIPv4Config represents an IPv4 configuration.
type BearerIPv4Config struct {
	// Address contains the IPv4 address.
	Address string `json:"address"`
	// DNS contains the suggested DNS servers.
	DNS []string `json:"dns"`
	// Gateway contains the gateway. In most cases,
	// this can be ignored as not all modems provide
	// this and the usual practice is to use an interface
	// route.
	Gateway string `json:"gateway"`
	// Method contains the interface configuration method.
	Method string `json:"method"`
	// MTU contains the maximum transmission unit.
	MTU Int `json:"mtu"`
	// Prefix contains the address prefix length.
	Prefix Int `json:"prefix"`
}

// BearerStatus represents the status of a bearer.
type BearerStatus struct {
	// Connected contains the connection status.
	Connected Boolean `json:"connected"`
	// Interface contains the interface name.
	Interface string `json:"interface"`
	// Suspended contains the suspended status.
	Suspended Boolean `json:"suspended"`
}

// Bearer represents a bearer.
type Bearer struct {
	// DBusPath contains the D-Bus path.
	DBusPath string `json:"dbus-path"`
	// IPv4Config contains the IPv4 configuration.
	IPv4Config BearerIPv4Config `json:"ipv4-config"`
	// Status contains the bearer status.
	Status BearerStatus `json:"status"`
}

// Connect connects the bearer.
func (b *Bearer) Connect(ctx context.Context) error {
	//nolint:gosec // We are not passing user input.
	cmd := exec.CommandContext(ctx, "mmcli",
		"--bearer="+b.DBusPath,
		"--connect",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		logging.FromContext(ctx).Error("Failed to run command", zap.String("output", string(output)))

		return fmt.Errorf("failed to connect bearer: %w", err)
	}

	return nil
}

// ConfigureInterface configures the bearer interface.
func (b *Bearer) ConfigureInterface(ctx context.Context) error {
	iface, err := ip.NewIface(
		b.Status.Interface,
		fmt.Sprintf("%s/%d", b.IPv4Config.Address, b.IPv4Config.Prefix),
		b.IPv4Config.Gateway,
		b.IPv4Config.DNS,
		int(b.IPv4Config.MTU),
	)
	if err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

	if err := iface.Reconcile(ctx); err != nil {
		return fmt.Errorf("failed to configure interface: %w", err)
	}

	return nil
}
