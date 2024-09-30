package ip

import (
	"context"
	"fmt"
	"net"
	"os/exec"
)

// Iface represents an interface.
type Iface struct {
	// Name contains the interface name.
	Name string `json:"name"`
	// AddressPrefix contains the address prefix.
	AddressPrefix net.IPNet `json:"address-prefix"`
	// Gateway contains the gateway.
	Gateway net.IP `json:"gateway"`
	// DNS contains the DNS servers.
	DNS []net.IP `json:"dns"`
	// MTU contains the maximum transmission unit.
	MTU int `json:"mtu"`
}

// NewIface creates a new interface.
func NewIface(name string, addressPrefix string, gateway string, dns []string, mtu int) (*Iface, error) {
	ip, addressPrefixNet, err := net.ParseCIDR(addressPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address prefix: %w", err)
	}

	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return nil, fmt.Errorf("failed to parse gateway: %w", err)
	}

	dnsIPs := make([]net.IP, 0, len(dns))
	for _, dnsStr := range dns {
		dnsIP := net.ParseIP(dnsStr)
		if dnsIP == nil {
			return nil, fmt.Errorf("failed to parse DNS: %w: %s", err, dnsStr)
		}

		dnsIPs = append(dnsIPs, dnsIP)
	}

	return &Iface{
		Name: name,
		AddressPrefix: net.IPNet{
			IP:   ip,
			Mask: addressPrefixNet.Mask,
		},
		Gateway: gatewayIP,
		DNS:     dnsIPs,
		MTU:     mtu,
	}, nil
}

// Reconcile reconciles the interface. Note that this
// currently only supports IPv4 and ignores DNS servers.
func (i *Iface) Reconcile(ctx context.Context) error {
	// Check if the interface already has the correct IP address.
	iface, err := net.InterfaceByName(i.Name)
	if err != nil {
		return fmt.Errorf("failed to get interface: %w", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return fmt.Errorf("failed to parse address: %w", err)
		}

		if i.AddressPrefix.IP.Equal(ip) {
			return nil
		}
	}

	// Set the IP address.
	if err := exec.CommandContext(ctx, "ip", "addr", "add", i.AddressPrefix.String(), "dev", i.Name).Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	// Set the gateway.
	if err := exec.CommandContext(ctx, "ip", "route", "add", "default", "via", i.Gateway.String()).Run(); err != nil {
		return fmt.Errorf("failed to set gateway: %w", err)
	}

	// Set the MTU.
	if err := exec.CommandContext(ctx, "ip", "link", "set", "dev", i.Name, "mtu", fmt.Sprint(i.MTU)).Run(); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	// Add route for the interface.
	if err := exec.CommandContext(ctx, "ip", "route", "add", i.AddressPrefix.String(), "dev", i.Name).Run(); err != nil {
		return fmt.Errorf("failed to add route: %w", err)
	}

	return nil
}
