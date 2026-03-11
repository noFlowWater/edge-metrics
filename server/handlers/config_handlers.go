package handlers

import (
	"database/sql"
	"edge-metrics-server/kubernetes"
	"edge-metrics-server/models"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func isSSRFSafe(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return !parsed.IsLoopback() && !parsed.IsLinkLocalUnicast() && !parsed.IsLinkLocalMulticast() && !parsed.IsMulticast()
}

// GetConfig handles GET /config/:device_id
func (h *Handler) GetConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Config request for device: %s", deviceID)

	config, err := h.devices.Get(deviceID)
	if err != nil {
		log.Printf("Error fetching config for %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch device configuration",
		})
		return
	}

	if config == nil {
		log.Printf("Device not found: %s", deviceID)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:    "Device not found",
			DeviceID: deviceID,
			Message:  "No configuration available for this device",
		})
		return
	}

	log.Printf("Returning config for %s: %s", deviceID, config.DeviceType)

	response := gin.H{
		"device_type": config.DeviceType,
		"port":        config.Port,
		"reload_port": config.ReloadPort,
	}

	if len(config.EnabledMetrics) > 0 {
		response["enabled_metrics"] = config.EnabledMetrics
	}

	for key, value := range config.ExtraConfig {
		response[key] = value
	}

	c.JSON(http.StatusOK, response)
}

// UpdateConfig handles PUT /config/:device_id
func (h *Handler) UpdateConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Update request for device: %s", deviceID)

	if err := kubernetes.ValidateDeviceID(deviceID); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_device_id",
			Message: err.Error(),
		})
		return
	}

	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		log.Printf("Invalid JSON for %s: %v", deviceID, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	config := models.DeviceConfig{}

	if deviceType, ok := rawData["device_type"].(string); ok {
		config.DeviceType = deviceType
	} else {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing required field",
			Message: "device_type is required",
		})
		return
	}

	if port, ok := rawData["port"].(float64); ok {
		config.Port = int(port)
	}
	if reloadPort, ok := rawData["reload_port"].(float64); ok {
		config.ReloadPort = int(reloadPort)
	}

	if metrics, ok := rawData["enabled_metrics"].([]interface{}); ok {
		for _, m := range metrics {
			if s, ok := m.(string); ok {
				config.EnabledMetrics = append(config.EnabledMetrics, s)
			}
		}
	}

	standardFields := map[string]bool{
		"device_type":     true,
		"port":            true,
		"reload_port":     true,
		"enabled_metrics": true,
		"ip_address":      true,
	}

	config.ExtraConfig = make(map[string]interface{})
	for key, value := range rawData {
		if !standardFields[key] {
			config.ExtraConfig[key] = value
		}
	}

	if config.Port == 0 {
		config.Port = 9100
	}
	if config.ReloadPort == 0 {
		config.ReloadPort = 9101
	}

	if ipAddress, ok := rawData["ip_address"].(string); ok {
		config.IPAddress = ipAddress
	}

	if config.IPAddress == "" {
		if existing, _ := h.devices.Get(deviceID); existing != nil {
			config.IPAddress = existing.IPAddress
		}
	} else {
		if !isValidIP(config.IPAddress) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_ip_address",
				Message: fmt.Sprintf("Invalid IP address format: %s", config.IPAddress),
			})
			return
		}
		if !isSSRFSafe(config.IPAddress) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "restricted_ip_address",
				Message: "IP address is in a restricted range (loopback, link-local, multicast)",
			})
			return
		}
	}

	created, err := h.devices.Upsert(deviceID, &config)
	if err != nil {
		log.Printf("Error upserting config for %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to save device configuration",
		})
		return
	}

	status := "updated"
	if created {
		status = "registered"
		log.Printf("Registered new device: %s", deviceID)
	} else {
		log.Printf("Updated config for device: %s", deviceID)
	}

	reloadTriggered := false
	if config.IPAddress != "" {
		reloadTriggered = TriggerDeviceReloadWithLogging(deviceID, config)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           status,
		"device_id":        deviceID,
		"reload_triggered": reloadTriggered,
	})
}

// CreateConfig handles POST /config/:device_id
func (h *Handler) CreateConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Create request for device: %s", deviceID)

	if err := kubernetes.ValidateDeviceID(deviceID); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_device_id",
			Message: err.Error(),
		})
		return
	}

	exists, err := h.devices.Exists(deviceID)
	if err != nil {
		log.Printf("Error checking device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to check device existence",
		})
		return
	}

	if exists {
		log.Printf("Device already exists: %s", deviceID)
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:    "Device already exists",
			DeviceID: deviceID,
			Message:  "Use PUT to update existing device",
		})
		return
	}

	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		log.Printf("Invalid JSON for %s: %v", deviceID, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	config := models.DeviceConfig{
		DeviceID: deviceID,
	}

	if deviceType, ok := rawData["device_type"].(string); ok {
		config.DeviceType = deviceType
	} else {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing required field",
			Message: "device_type is required",
		})
		return
	}

	if port, ok := rawData["port"].(float64); ok {
		config.Port = int(port)
	}
	if reloadPort, ok := rawData["reload_port"].(float64); ok {
		config.ReloadPort = int(reloadPort)
	}

	if metrics, ok := rawData["enabled_metrics"].([]interface{}); ok {
		for _, m := range metrics {
			if s, ok := m.(string); ok {
				config.EnabledMetrics = append(config.EnabledMetrics, s)
			}
		}
	}

	standardFields := map[string]bool{
		"device_type":     true,
		"port":            true,
		"reload_port":     true,
		"enabled_metrics": true,
		"ip_address":      true,
	}

	config.ExtraConfig = make(map[string]interface{})
	for key, value := range rawData {
		if !standardFields[key] {
			config.ExtraConfig[key] = value
		}
	}

	if ipAddress, ok := rawData["ip_address"].(string); ok {
		config.IPAddress = ipAddress
	}

	if config.IPAddress == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "ip_address_required",
			Message: "Device IP address must be specified in configuration",
		})
		return
	}

	if !isValidIP(config.IPAddress) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_ip_address",
			Message: fmt.Sprintf("Invalid IP address format: %s", config.IPAddress),
		})
		return
	}

	if !isSSRFSafe(config.IPAddress) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "restricted_ip_address",
			Message: "IP address is in a restricted range (loopback, link-local, multicast)",
		})
		return
	}

	if config.Port == 0 {
		config.Port = 9100
	}
	if config.ReloadPort == 0 {
		config.ReloadPort = 9101
	}

	err = h.devices.Create(&config)
	if err != nil {
		log.Printf("Error creating config for %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to create device configuration",
		})
		return
	}

	log.Printf("Created new device: %s", deviceID)
	c.JSON(http.StatusCreated, models.UpdateResponse{
		Status:   "created",
		DeviceID: deviceID,
	})
}

// DeleteConfig handles DELETE /config/:device_id
func (h *Handler) DeleteConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Delete request for device: %s", deviceID)

	err := h.devices.Delete(deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Device not found for delete: %s", deviceID)
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:    "Device not found",
				DeviceID: deviceID,
			})
			return
		}
		log.Printf("Error deleting config for %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to delete device configuration",
		})
		return
	}

	log.Printf("Deleted device: %s", deviceID)
	c.JSON(http.StatusOK, models.UpdateResponse{
		Status:   "deleted",
		DeviceID: deviceID,
	})
}

// ListConfigs handles GET /config
func (h *Handler) ListConfigs(c *gin.Context) {
	log.Printf("List all configs request")

	devices, err := h.devices.List()
	if err != nil {
		log.Printf("Error fetching configs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch configurations",
		})
		return
	}

	configs := make([]gin.H, 0)
	for _, device := range devices {
		config := gin.H{
			"device_id":   device.DeviceID,
			"device_type": device.DeviceType,
			"port":        device.Port,
			"reload_port": device.ReloadPort,
		}

		if len(device.EnabledMetrics) > 0 {
			config["enabled_metrics"] = device.EnabledMetrics
		}

		for key, value := range device.ExtraConfig {
			config[key] = value
		}

		configs = append(configs, config)
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"total":   len(configs),
	})
}

// PatchConfig handles PATCH /config/:device_id
func (h *Handler) PatchConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	log.Printf("Patch request for device: %s", deviceID)

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
			Message:  "Use POST or PUT to create new device",
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
		if val == nil {
			existing.DeviceType = ""
		} else if s, ok := val.(string); ok {
			existing.DeviceType = s
		}
	}
	if val, exists := patchData["port"]; exists {
		if val == nil {
			existing.Port = 9100
		} else if f, ok := val.(float64); ok {
			existing.Port = int(f)
		}
	}
	if val, exists := patchData["reload_port"]; exists {
		if val == nil {
			existing.ReloadPort = 9101
		} else if f, ok := val.(float64); ok {
			existing.ReloadPort = int(f)
		}
	}
	if val, exists := patchData["enabled_metrics"]; exists {
		if val == nil {
			existing.EnabledMetrics = nil
		} else if metrics, ok := val.([]interface{}); ok {
			existing.EnabledMetrics = nil
			for _, m := range metrics {
				if s, ok := m.(string); ok {
					existing.EnabledMetrics = append(existing.EnabledMetrics, s)
				}
			}
		}
	}

	if val, exists := patchData["ip_address"]; exists {
		if val == nil {
			// null value - keep existing IP
		} else if s, ok := val.(string); ok {
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

	standardFields := map[string]bool{
		"device_type":     true,
		"port":            true,
		"reload_port":     true,
		"enabled_metrics": true,
		"ip_address":      true,
	}

	if existing.ExtraConfig == nil {
		existing.ExtraConfig = make(map[string]interface{})
	}

	for key, value := range patchData {
		if !standardFields[key] {
			if value == nil {
				delete(existing.ExtraConfig, key)
			} else {
				existing.ExtraConfig[key] = value
			}
		}
	}

	err = h.devices.Update(deviceID, existing)
	if err != nil {
		log.Printf("Error updating config for %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to update device configuration",
		})
		return
	}

	reloadTriggered := false
	if existing.IPAddress != "" {
		reloadTriggered = TriggerDeviceReloadWithLogging(deviceID, *existing)
	}

	log.Printf("Patched config for device: %s", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"status":           "patched",
		"device_id":        deviceID,
		"reload_triggered": reloadTriggered,
	})
}
