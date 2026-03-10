package models

// DeviceConfig represents the configuration for a device
type DeviceConfig struct {
	DeviceID       string                 `json:"-"`
	DeviceType     string                 `json:"device_type"`
	Port           int                    `json:"port,omitempty"`
	ReloadPort     int                    `json:"reload_port,omitempty"`
	EnabledMetrics []string               `json:"enabled_metrics,omitempty"`
	ExtraConfig    map[string]interface{} `json:"-"` // Device-specific config (shelly, jetson, ina260, etc.)
	IPAddress      string                 `json:"ip_address"`
}

// DeviceStatus represents a device with its health status
type DeviceStatus struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	IPAddress  string `json:"ip_address,omitempty"`
	Port       int    `json:"port"`
	ReloadPort int    `json:"reload_port"`
	Status     string `json:"status"` // healthy, unhealthy, unreachable
	LastSeen   string `json:"last_seen,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DevicesListResponse represents the response for listing all devices
type DevicesListResponse struct {
	Devices   []DeviceStatus `json:"devices"`
	Total     int            `json:"total"`
	Healthy   int            `json:"healthy"`
	Unhealthy int            `json:"unhealthy"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error    string `json:"error"`
	DeviceID string `json:"device_id,omitempty"`
	Message  string `json:"message,omitempty"`
}

// UpdateResponse represents a successful update response
type UpdateResponse struct {
	Status   string `json:"status"`
	DeviceID string `json:"device_id"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}
