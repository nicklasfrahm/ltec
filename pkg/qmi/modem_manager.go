package qmi

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/nicklasfrahm-dev/appkit/logging"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

const (
	// DefaultDevice is the default modem device path.
	DefaultDevice = "/dev/cdc-wdm0"
	// DefaultAPN is the default access point name.
	DefaultAPN = "bredband.oister.dk"
)

// ModemStatus the status of the modem.
type ModemStatus struct {
	// iface is a gauge to monitor the interface connection status.
	iface prometheus.Gauge
	// internet is a gauge to monitor the internet connection status.
	internet prometheus.Gauge
}

// ModemManager represents a component that manages a modem.
type ModemManager struct {
	// device is the modem device path, e.g. `/dev/cdc-wdm0`.
	device string
	// apn is the access point name for the modem, e.g. `apn.example.com`.
	apn string
	// logger is the logger for the modem.
	logger *zap.Logger
	// status is the modem status.
	status *ModemStatus
}

// NewModemManager creates a new Modem instance
func NewModemManager(ctx context.Context, registry *prometheus.Registry) *ModemManager {
	logger := logging.FromContext(ctx)

	modem := &ModemManager{
		device: DefaultDevice,
		apn:    DefaultAPN,
		logger: logger,
		status: &ModemStatus{
			iface: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "wwand_interface_connection_status",
				Help: "Interface connection status (1 = connected, 0 = disconnected)",
			}),
			internet: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "wwand_internet_connection_status",
				Help: "Internet connection status (1 = connected, 0 = disconnected)",
			}),
		},
	}

	if registry != nil {
		registry.MustRegister(modem.status.iface)
		registry.MustRegister(modem.status.internet)

		modem.status.iface.Set(0)
		modem.status.internet.Set(0)
	}

	return modem
}

// Start starts the ModemManager component.
func (m *ModemManager) Start(ctx context.Context) error {
	if err := m.configure(); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// TODO: Finish implementation.

	return nil
}

// Stop stops the ModemManager component.
func (m *ModemManager) Stop(ctx context.Context) error {
	// TODO: Finish implementation.

	return nil
}

// configure loads the modem configuration.
func (m *ModemManager) configure() error {
	device := os.Getenv("WWAND_DEVICE")
	if device != "" {
		m.device = device
	}

	apn := os.Getenv("WWAND_APN")
	if apn != "" {
		m.apn = apn
	}

	return nil
}

// runQMICommand runs a qmicli command with the given arguments.
func (m *ModemManager) runQMICommand(args ...string) (string, error) {
	cmdArgs := append([]string{fmt.Sprintf("--device=%s", m.device)}, args...)

	cmd := exec.Command("qmicli", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		m.logger.Error("Failed to run qmicli command",
			zap.String("output", strings.TrimSpace(string(out))),
			zap.Error(err),
		)

		return "", fmt.Errorf("failed to run qmicli command: %v", err)
	}

	return string(out), err
}

// ConfigureModem configures the modem to connect to the APN
func (m *ModemManager) ConfigureModem() error {
	_, err := m.runQMICommand("--device-open-proxy", fmt.Sprintf("--wds-start-network=apn=%s,ip-type=4", m.apn), "--client-no-release-cid")
	if err != nil {
		return fmt.Errorf("failed to configure modem: %v", err)
	}
	return nil
}

// CheckConnection checks the modem's connection status
func (m *ModemManager) CheckConnection() bool {
	output, err := m.runQMICommand("--wds-get-packet-service-status")
	if err != nil {
		log.Printf("Error checking connection: %v", err)
		return false
	}
	return strings.Contains(output, "connected")
}

// ReconfigureInterface reconfigures network settings when IP changes
func (m *ModemManager) ReconfigureInterface() error {
	output, err := m.runQMICommand("--wds-get-current-settings")
	if err != nil {
		return fmt.Errorf("failed to get current settings: %v", err)
	}
	ip, gateway := parseIPSettings(output)
	if ip != "" && gateway != "" {
		if err := configureIPAndRoute(ip, gateway); err != nil {
			return fmt.Errorf("failed to configure IP/route: %v", err)
		}
	}
	return nil
}

// Parse IP and Gateway from qmicli output (simplified example)
func parseIPSettings(_ string) (string, string) {
	ip := "192.168.1.100"
	gateway := "192.168.1.1"
	return ip, gateway
}

// Configure IP and route using system commands
func configureIPAndRoute(ip string, gateway string) error {
	_, err := exec.Command("ip", "addr", "add", ip+"/24", "dev", "wwan0").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure IP: %v", err)
	}
	_, err = exec.Command("ip", "route", "add", "default", "via", gateway, "dev", "wwan0").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure route: %v", err)
	}
	return nil
}

// MonitorConnection listens for network link changes and handles disconnections
func (m *ModemManager) MonitorConnection() {
	for {
		linkDown := waitForLinkDown()
		if linkDown {
			m.status.iface.Set(0)
			log.Println("Connection lost, attempting to reconnect...")
			if err := m.ConfigureModem(); err != nil {
				log.Printf("Reconnection failed: %v", err)
			} else {
				m.ReconfigureInterface()
			}
		}
	}
}

// waitForLinkDown uses a netlink-based event listener to detect link changes
func waitForLinkDown() bool {
	// Using netlink to listen for changes on the interface
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		log.Fatalf("Failed to create netlink socket: %v", err)
	}
	defer unix.Close(fd)

	// Prepare to listen for link state changes (simplified version)
	for {
		buf := make([]byte, 4096)
		_, _, err := unix.Recvfrom(fd, buf, 0)
		if err != nil {
			log.Printf("Error receiving netlink message: %v", err)
			continue
		}

		// Simplified: Detect link down events (this needs to be expanded for real-world use)
		if strings.Contains(string(buf), "IFLA_OPERSTATE_DOWN") {
			return true
		}
	}
}
