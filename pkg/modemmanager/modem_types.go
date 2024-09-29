package modemmanager

// ModemStatusResponse represents a modem status response.
type ModemStatusResponse struct {
	// Modem contains the status of the requested modem.
	Modem ModemStatus `json:"modem"`
}

// ModemStatus represents the top-level structure for modem information.
type ModemStatus struct {
	// DBusPath is the D-Bus path of the modem.
	DBusPath string `json:"dbus-path"`
	// Generic contains generic modem information.
	Generic ModemStatusGeneric `json:"generic"`
}

// ModemStatusGeneric represents generic modem information.
type ModemStatusGeneric struct {
	// Bearers is a list of bearer paths.
	Bearers []string `json:"bearers"`
	// State is the state of the modem.
	State string `json:"state"`
}
