package handlers

import (
	"fmt"
	"net/http"
	"os"

	"edge-metrics-server/kubernetes"
	"edge-metrics-server/models"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SyncKubernetesRequest represents the request body for sync operation
type SyncKubernetesRequest struct {
	Namespace string `json:"namespace"`
}

// SyncKubernetes handles POST /kubernetes/sync
func (h *Handler) SyncKubernetes(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	if h.syncer != nil {
		result, err := h.syncer.TriggerSync()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Sync failed",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// Fallback: direct sync without syncer
	var req SyncKubernetesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Namespace = "monitoring"
	}
	if req.Namespace == "" {
		req.Namespace = "monitoring"
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8081"
	}

	result, err := kubernetes.SyncDevices(req.Namespace, serverURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetManifests handles GET /kubernetes/manifests
func (h *Handler) GetManifests(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "monitoring")

	configs, err := h.devices.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get device configs",
			"message": err.Error(),
		})
		return
	}

	var healthyDevices []models.DeviceConfig
	for _, config := range configs {
		if IsDeviceHealthy(config) {
			healthyDevices = append(healthyDevices, config)
		}
	}

	if len(healthyDevices) == 0 {
		c.String(http.StatusOK, "# No healthy devices to generate manifests\n")
		return
	}

	yaml := generateManifests(namespace, healthyDevices)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, yaml)
}

// GetKubernetesStatus handles GET /kubernetes/status
func (h *Handler) GetKubernetesStatus(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	namespace := c.DefaultQuery("namespace", "monitoring")

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8081"
	}

	status, err := kubernetes.GetSyncStatus(namespace, serverURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SyncSingleDevice handles POST /kubernetes/sync/:device_id
func (h *Handler) SyncSingleDevice(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	deviceID := c.Param("device_id")
	namespace := c.DefaultQuery("namespace", "monitoring")

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8081"
	}

	result, err := kubernetes.SyncSingleDevice(namespace, deviceID, serverURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDeviceResources handles GET /kubernetes/resources/:device_id
func (h *Handler) GetDeviceResources(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	deviceID := c.Param("device_id")
	namespace := c.DefaultQuery("namespace", "monitoring")

	resources, err := kubernetes.GetDeviceResources(namespace, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get resources",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resources)
}

// DeleteDeviceResources handles DELETE /kubernetes/resources/:device_id
func (h *Handler) DeleteDeviceResources(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	deviceID := c.Param("device_id")
	namespace := c.DefaultQuery("namespace", "monitoring")

	result, err := kubernetes.DeleteDeviceResources(namespace, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Delete failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetKubernetesHealth handles GET /kubernetes/health
func (h *Handler) GetKubernetesHealth(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "monitoring")

	health, err := kubernetes.CheckHealth(namespace)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

// CleanupKubernetes handles DELETE /kubernetes/cleanup
func (h *Handler) CleanupKubernetes(c *gin.Context) {
	if !kubernetes.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Kubernetes client not initialized",
			"message": "Server not running in Kubernetes environment or kubeconfig not found",
		})
		return
	}

	namespace := c.DefaultQuery("namespace", "monitoring")

	services, endpoints, err := kubernetes.CleanupAllResources(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Cleanup failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "cleaned",
		"deleted_services":  services,
		"deleted_endpoints": endpoints,
		"namespace":         namespace,
	})
}

// generateManifests generates Kubernetes YAML manifests for healthy devices
func generateManifests(namespace string, devices []models.DeviceConfig) string {
	yaml := fmt.Sprintf("# Kubernetes manifests for edge devices\n")
	yaml += fmt.Sprintf("# Generated for namespace: %s\n\n", namespace)

	for _, device := range devices {
		if device.IPAddress == "" {
			continue
		}

		serviceName := fmt.Sprintf("edge-device-%s", device.DeviceID)

		service := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: namespace,
				Labels: map[string]string{
					"app":         "edge-exporter",
					"device_id":   device.DeviceID,
					"device_type": device.DeviceType,
					"managed_by":  "edge-metrics-server",
				},
			},
			Spec: corev1.ServiceSpec{
				ClusterIP: "None",
				Ports: []corev1.ServicePort{
					{
						Name:       "metrics",
						Port:       int32(device.Port),
						TargetPort: intstr.FromInt(device.Port),
						Protocol:   corev1.ProtocolTCP,
					},
				},
			},
		}

		yaml += "---\n"
		yaml += fmt.Sprintf("apiVersion: %s\n", service.APIVersion)
		yaml += fmt.Sprintf("kind: %s\n", service.Kind)
		yaml += "metadata:\n"
		yaml += fmt.Sprintf("  name: %s\n", service.Name)
		yaml += fmt.Sprintf("  namespace: %s\n", service.Namespace)
		yaml += "  labels:\n"
		for k, v := range service.Labels {
			yaml += fmt.Sprintf("    %s: %s\n", k, v)
		}
		yaml += "spec:\n"
		yaml += fmt.Sprintf("  clusterIP: %s\n", service.Spec.ClusterIP)
		yaml += "  ports:\n"
		for _, port := range service.Spec.Ports {
			yaml += fmt.Sprintf("  - name: %s\n", port.Name)
			yaml += fmt.Sprintf("    port: %d\n", port.Port)
			yaml += fmt.Sprintf("    targetPort: %d\n", port.TargetPort.IntVal)
			yaml += fmt.Sprintf("    protocol: %s\n", port.Protocol)
		}

		yaml += "---\n"
		yaml += "apiVersion: v1\n"
		yaml += "kind: Endpoints\n"
		yaml += "metadata:\n"
		yaml += fmt.Sprintf("  name: %s\n", serviceName)
		yaml += fmt.Sprintf("  namespace: %s\n", namespace)
		yaml += "  labels:\n"
		yaml += fmt.Sprintf("    app: edge-exporter\n")
		yaml += fmt.Sprintf("    device_id: %s\n", device.DeviceID)
		yaml += fmt.Sprintf("    managed_by: edge-metrics-server\n")
		yaml += "subsets:\n"
		yaml += "- addresses:\n"
		yaml += fmt.Sprintf("  - ip: %s\n", device.IPAddress)
		yaml += "  ports:\n"
		yaml += fmt.Sprintf("  - name: metrics\n")
		yaml += fmt.Sprintf("    port: %d\n", device.Port)
		yaml += fmt.Sprintf("    protocol: TCP\n")
		yaml += "\n"
	}

	return yaml
}
