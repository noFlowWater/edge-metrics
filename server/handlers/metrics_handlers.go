package handlers

import (
	"edge-metrics-server/models"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMetricsSummary handles GET /metrics/summary
func (h *Handler) GetMetricsSummary(c *gin.Context) {
	log.Printf("Metrics summary request")

	devices, err := h.devices.List()
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to fetch devices",
		})
		return
	}

	typeCount := make(map[string]int)
	for _, device := range devices {
		typeCount[device.DeviceType]++
	}

	client := &http.Client{Timeout: 2 * time.Second}

	// Per-device health result: true = healthy, false = unhealthy
	results := make([]bool, len(devices))
	var wg sync.WaitGroup
	wg.Add(len(devices))
	for i, device := range devices {
		go func(idx int, dev models.DeviceConfig) {
			defer wg.Done()
			if dev.IPAddress == "" {
				results[idx] = false
				return
			}
			healthURL := fmt.Sprintf("http://%s:%d/health", dev.IPAddress, dev.ReloadPort)
			resp, err := client.Get(healthURL)
			if err != nil {
				results[idx] = false
				return
			}
			resp.Body.Close()
			results[idx] = resp.StatusCode == http.StatusOK
		}(i, device)
	}
	wg.Wait()

	healthy := 0
	unhealthy := 0
	for _, ok := range results {
		if ok {
			healthy++
		} else {
			unhealthy++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":          len(devices),
		"healthy":        healthy,
		"unhealthy":      unhealthy,
		"by_device_type": typeCount,
	})
}
