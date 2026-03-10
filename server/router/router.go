package router

import (
	"edge-metrics-server/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// Config routes
	r.GET("/config", h.ListConfigs)
	r.GET("/config/:device_id", h.GetConfig)
	r.POST("/config/:device_id", h.CreateConfig)
	r.PUT("/config/:device_id", h.UpdateConfig)
	r.PATCH("/config/:device_id", h.PatchConfig)
	r.DELETE("/config/:device_id", h.DeleteConfig)

	// Device routes
	r.GET("/devices", h.ListDevices)
	r.POST("/devices/reload", h.ReloadAllDevices)
	r.GET("/devices/:device_id/status", h.GetDeviceStatus)
	r.PATCH("/devices/:device_id", h.PatchDevice)
	r.GET("/devices/:device_id/local-config", h.GetDeviceLocalConfig)
	r.POST("/devices/:device_id/reload", h.ReloadDevice)

	// Metrics routes
	r.GET("/metrics/summary", h.GetMetricsSummary)

	// Kubernetes routes
	r.GET("/kubernetes/status", h.GetKubernetesStatus)
	r.GET("/kubernetes/health", h.GetKubernetesHealth)
	r.POST("/kubernetes/sync", h.SyncKubernetes)
	r.POST("/kubernetes/sync/:device_id", h.SyncSingleDevice)
	r.GET("/kubernetes/manifests", h.GetManifests)
	r.GET("/kubernetes/resources/:device_id", h.GetDeviceResources)
	r.DELETE("/kubernetes/resources/:device_id", h.DeleteDeviceResources)
	r.DELETE("/kubernetes/cleanup", h.CleanupKubernetes)

	// Health route
	r.GET("/health", h.Health)
}
