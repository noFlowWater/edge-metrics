package handlers

import (
	"edge-metrics-server/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ListDevices handles GET /devices
func (h *Handler) ListDevices(c *gin.Context) {
	log.Printf("List devices request")

	devices, err := h.devices.List()
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch devices",
		})
		return
	}

	deviceStatuses := make([]models.DeviceStatus, len(devices))
	healthy := 0
	unhealthy := 0

	var wg sync.WaitGroup
	wg.Add(len(devices))
	for i, device := range devices {
		go func(idx int, dev models.DeviceConfig) {
			defer wg.Done()
			deviceStatuses[idx] = CheckDeviceHealth(dev)
		}(i, device)
	}
	wg.Wait()

	for _, status := range deviceStatuses {
		if status.Status == "healthy" {
			healthy++
		} else {
			unhealthy++
		}
	}

	c.JSON(http.StatusOK, models.DevicesListResponse{
		Devices:   deviceStatuses,
		Total:     len(devices),
		Healthy:   healthy,
		Unhealthy: unhealthy,
	})
}

// GetDeviceStatus handles GET /devices/:device_id/status
func (h *Handler) GetDeviceStatus(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Device status request for: %s", deviceID)

	device, err := h.devices.Get(deviceID)
	if err != nil {
		log.Printf("Error fetching device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch device",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:    "Device not found",
			DeviceID: deviceID,
		})
		return
	}

	status := CheckDeviceHealth(*device)
	c.JSON(http.StatusOK, status)
}

// ReloadDevice handles POST /devices/:device_id/reload
func (h *Handler) ReloadDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Reload request for device: %s", deviceID)

	device, err := h.devices.Get(deviceID)
	if err != nil {
		log.Printf("Error fetching device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch device",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:    "Device not found",
			DeviceID: deviceID,
		})
		return
	}

	success, errMsg := TriggerDeviceReload(*device)
	if !success {
		log.Printf("Failed to trigger reload for %s: %s", deviceID, errMsg)
		statusCode := http.StatusServiceUnavailable
		if errMsg == "No IP address" {
			statusCode = http.StatusBadRequest
		} else if len(errMsg) >= 4 && errMsg[:4] == "HTTP" {
			statusCode = http.StatusBadGateway
		}
		c.JSON(statusCode, gin.H{
			"status":    "failed",
			"device_id": deviceID,
			"error":     errMsg,
		})
		return
	}

	log.Printf("Reload triggered for device: %s", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"status":    "reloaded",
		"device_id": deviceID,
	})
}

// ReloadAllDevices handles POST /devices/reload
func (h *Handler) ReloadAllDevices(c *gin.Context) {
	log.Printf("Reload all devices request")

	devices, err := h.devices.List()
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch devices",
		})
		return
	}

	results := make([]gin.H, 0)
	success := 0
	failed := 0

	for _, device := range devices {
		result := gin.H{
			"device_id": device.DeviceID,
		}

		reloadSuccess, errMsg := TriggerDeviceReload(device)
		if reloadSuccess {
			result["status"] = "reloaded"
			success++
		} else {
			if errMsg == "No IP address" {
				result["status"] = "skipped"
			} else {
				result["status"] = "failed"
			}
			result["error"] = errMsg
			failed++
		}

		results = append(results, result)
	}

	log.Printf("Reload all: %d success, %d failed", success, failed)
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(devices),
		"success": success,
		"failed":  failed,
	})
}

// GetDeviceLocalConfig handles GET /devices/:device_id/local-config
func (h *Handler) GetDeviceLocalConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Local config request for device: %s", deviceID)

	device, err := h.devices.Get(deviceID)
	if err != nil {
		log.Printf("Error fetching device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch device",
		})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:    "Device not found",
			DeviceID: deviceID,
		})
		return
	}

	if device.IPAddress == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:    "No IP address",
			DeviceID: deviceID,
			Message:  "Device has no IP address configured",
		})
		return
	}

	if !isSSRFSafe(device.IPAddress) {
		log.Printf("SSRF blocked: device %s has disallowed IP %s", deviceID, device.IPAddress)
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:    "Forbidden IP range",
			DeviceID: deviceID,
			Message:  "Device IP address is in a restricted range",
		})
		return
	}

	configURL := fmt.Sprintf("http://%s:%d/config", device.IPAddress, device.ReloadPort)
	log.Printf("Fetching local config from: %s", configURL)

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(configURL)
	if err != nil {
		log.Printf("Failed to fetch local config from %s: %v", configURL, err)
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:    "Device unreachable",
			DeviceID: deviceID,
			Message:  fmt.Sprintf("Failed to connect to device: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Device returned non-OK status: %d", resp.StatusCode)
		c.JSON(http.StatusBadGateway, models.ErrorResponse{
			Error:    "Device error",
			DeviceID: deviceID,
			Message:  fmt.Sprintf("Device returned HTTP %d", resp.StatusCode),
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

// PatchDevice handles PATCH /devices/:device_id
func (h *Handler) PatchDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Patch device basic info request for: %s", deviceID)

	existing, err := h.devices.Get(deviceID)
	if err != nil {
		log.Printf("Error fetching device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch device",
		})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:    "Device not found",
			DeviceID: deviceID,
		})
		return
	}

	var patchData map[string]interface{}
	if err := c.ShouldBindJSON(&patchData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	if val, exists := patchData["device_type"]; exists {
		if s, ok := val.(string); ok && s != "" {
			existing.DeviceType = s
		}
	}

	if val, exists := patchData["ip_address"]; exists {
		if s, ok := val.(string); ok {
			if s == "" {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "invalid_ip_address",
					Message: "IP address cannot be empty",
				})
				return
			}
			if !isValidIP(s) {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "invalid_ip_address",
					Message: fmt.Sprintf("Invalid IP address format: %s", s),
				})
				return
			}
			if !isSSRFSafe(s) {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "restricted_ip_address",
					Message: "IP address is in a restricted range (loopback, link-local, multicast)",
				})
				return
			}
			existing.IPAddress = s
		}
	}

	if val, exists := patchData["port"]; exists {
		if f, ok := val.(float64); ok {
			existing.Port = int(f)
		}
	}

	if val, exists := patchData["reload_port"]; exists {
		if f, ok := val.(float64); ok {
			existing.ReloadPort = int(f)
		}
	}

	err = h.devices.Update(deviceID, existing)
	if err != nil {
		log.Printf("Error updating device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to update device",
		})
		return
	}

	log.Printf("Updated basic info for device: %s (reload NOT triggered)", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"status":    "updated",
		"device_id": deviceID,
		"message":   "Device basic information updated (reload not triggered)",
	})
}
