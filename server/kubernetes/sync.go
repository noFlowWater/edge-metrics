package kubernetes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"edge-metrics-server/models"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	DeviceID string `json:"device_id"`
	Service  string `json:"service,omitempty"`
	Status   string `json:"status"` // created, updated, failed
	Error    string `json:"error,omitempty"`
}

// SyncResponse represents the response from a sync operation
type SyncResponse struct {
	Status       string       `json:"status"`
	Created      []SyncResult `json:"created"`
	Updated      []SyncResult `json:"updated"`
	Deleted      []SyncResult `json:"deleted"`
	Failed       []SyncResult `json:"failed"`
	TotalHealthy int          `json:"total_healthy"`
}

// SyncDevices synchronizes healthy devices to Kubernetes
func SyncDevices(namespace, serverURL string) (*SyncResponse, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	// Get healthy devices from the API
	devices, err := getHealthyDevices(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy devices: %w", err)
	}

	response := &SyncResponse{
		Status:       "synced",
		Created:      []SyncResult{},
		Updated:      []SyncResult{},
		Deleted:      []SyncResult{},
		Failed:       []SyncResult{},
		TotalHealthy: len(devices),
	}

	// Track existing services in K8s
	existingServices, err := ListEdgeServices(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing services: %w", err)
	}
	existingMap := make(map[string]bool)
	for _, svc := range existingServices {
		existingMap[svc] = true
	}

	// Create or update services for healthy devices
	for _, device := range devices {
		if device.IPAddress == "" {
			response.Failed = append(response.Failed, SyncResult{
				DeviceID: device.DeviceID,
				Status:   "failed",
				Error:    "no IP address registered",
			})
			continue
		}

		serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(device.DeviceID))
		wasExisting := existingMap[serviceName]
		delete(existingMap, serviceName) // Mark as processed

		// Create/Update Service
		err := CreateOrUpdateService(namespace, device.DeviceID, device.DeviceType, device.Port)
		if err != nil {
			response.Failed = append(response.Failed, SyncResult{
				DeviceID: device.DeviceID,
				Service:  serviceName,
				Status:   "failed",
				Error:    fmt.Sprintf("service: %v", err),
			})
			continue
		}

		// Create/Update Endpoints
		err = CreateOrUpdateEndpoints(namespace, device.DeviceID, device.IPAddress, device.Port)
		if err != nil {
			response.Failed = append(response.Failed, SyncResult{
				DeviceID: device.DeviceID,
				Service:  serviceName,
				Status:   "failed",
				Error:    fmt.Sprintf("endpoints: %v", err),
			})
			continue
		}

		result := SyncResult{
			DeviceID: device.DeviceID,
			Service:  serviceName,
		}

		if wasExisting {
			result.Status = "updated"
			response.Updated = append(response.Updated, result)
		} else {
			result.Status = "created"
			response.Created = append(response.Created, result)
		}
	}

	// Delete services that are no longer healthy
	for serviceName := range existingMap {
		// Extract device_id from service name (edge-device-{device_id})
		deviceID := serviceName[len("edge-device-"):]

		err := DeleteService(namespace, deviceID)
		if err != nil {
			response.Failed = append(response.Failed, SyncResult{
				DeviceID: deviceID,
				Service:  serviceName,
				Status:   "failed",
				Error:    fmt.Sprintf("delete service: %v", err),
			})
			continue
		}

		err = DeleteEndpoints(namespace, deviceID)
		if err != nil {
			response.Failed = append(response.Failed, SyncResult{
				DeviceID: deviceID,
				Service:  serviceName,
				Status:   "failed",
				Error:    fmt.Sprintf("delete endpoints: %v", err),
			})
			continue
		}

		response.Deleted = append(response.Deleted, SyncResult{
			DeviceID: deviceID,
			Service:  serviceName,
			Status:   "deleted",
		})
	}

	return response, nil
}

// getHealthyDevices fetches healthy devices from the server API
func getHealthyDevices(serverURL string) ([]models.DeviceStatus, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/devices", serverURL)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var listResponse models.DevicesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, err
	}

	// Filter only healthy devices
	var healthyDevices []models.DeviceStatus
	for _, device := range listResponse.Devices {
		if device.Status == "healthy" {
			healthyDevices = append(healthyDevices, device)
		}
	}

	return healthyDevices, nil
}

// SyncStatusResponse represents overall sync status
type SyncStatusResponse struct {
	KubernetesEnabled  bool                  `json:"kubernetes_enabled"`
	Namespace          string                `json:"namespace"`
	TotalK8sResources  int                   `json:"total_k8s_resources"`
	TotalDevices       int                   `json:"total_registered_devices"`
	Synced             int                   `json:"synced"`
	Unsynced           int                   `json:"unsynced"`
	Resources          []DeviceResourceInfo  `json:"resources"`
}

// DeviceResourceInfo represents K8s resource info for a device
type DeviceResourceInfo struct {
	DeviceID        string `json:"device_id"`
	ServiceExists   bool   `json:"service_exists"`
	EndpointsExists bool   `json:"endpoints_exists"`
}

// DeviceResourceDetail represents detailed K8s resource info for a device
type DeviceResourceDetail struct {
	DeviceID         string              `json:"device_id"`
	Service          ServiceInfo         `json:"service"`
	Endpoints        EndpointsInfo       `json:"endpoints"`
	PrometheusTarget string              `json:"prometheus_target"`
}

// ServiceInfo represents Service resource info
type ServiceInfo struct {
	Name      string     `json:"name"`
	Exists    bool       `json:"exists"`
	ClusterIP string     `json:"cluster_ip"`
	Ports     []PortInfo `json:"ports"`
}

// PortInfo represents port information
type PortInfo struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
}

// EndpointsInfo represents Endpoints resource info
type EndpointsInfo struct {
	Name              string   `json:"name"`
	Exists            bool     `json:"exists"`
	ReadyAddresses    []string `json:"ready_addresses"`
	NotReadyAddresses []string `json:"not_ready_addresses"`
}

// GetSyncStatus returns the current sync status
func GetSyncStatus(namespace, serverURL string) (*SyncStatusResponse, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	// Get all devices from API
	devices, err := getAllDevices(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	// List existing services
	services, err := ListEdgeServices(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	serviceMap := make(map[string]bool)
	for _, svc := range services {
		serviceMap[svc] = true
	}

	// List existing endpoints
	endpoints, err := ListEdgeEndpoints(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}
	endpointMap := make(map[string]bool)
	for _, ep := range endpoints {
		endpointMap[ep] = true
	}

	response := &SyncStatusResponse{
		KubernetesEnabled: true,
		Namespace:         namespace,
		TotalK8sResources: len(services),
		TotalDevices:      len(devices),
		Resources:         []DeviceResourceInfo{},
	}

	synced := 0
	for _, device := range devices {
		serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(device.DeviceID))
		serviceExists := serviceMap[serviceName]
		endpointsExists := endpointMap[serviceName]

		if serviceExists && endpointsExists {
			synced++
		}

		response.Resources = append(response.Resources, DeviceResourceInfo{
			DeviceID:        device.DeviceID,
			ServiceExists:   serviceExists,
			EndpointsExists: endpointsExists,
		})
	}

	response.Synced = synced
	response.Unsynced = len(devices) - synced

	return response, nil
}

// SyncSingleDevice synchronizes a single device to Kubernetes
func SyncSingleDevice(namespace, deviceID, serverURL string) (*SyncResult, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	// Get device info from API
	device, err := getDeviceByID(serverURL, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if device.Status != "healthy" {
		return &SyncResult{
			DeviceID: deviceID,
			Status:   "failed",
			Error:    "device is not healthy",
		}, nil
	}

	if device.IPAddress == "" {
		return &SyncResult{
			DeviceID: deviceID,
			Status:   "failed",
			Error:    "no IP address registered",
		}, nil
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))

	// Check if service exists
	services, _ := ListEdgeServices(namespace)
	wasExisting := false
	for _, svc := range services {
		if svc == serviceName {
			wasExisting = true
			break
		}
	}

	// Create/Update Service
	err = CreateOrUpdateService(namespace, device.DeviceID, device.DeviceType, device.Port)
	if err != nil {
		return &SyncResult{
			DeviceID: deviceID,
			Service:  serviceName,
			Status:   "failed",
			Error:    fmt.Sprintf("service: %v", err),
		}, nil
	}

	// Create/Update Endpoints
	err = CreateOrUpdateEndpoints(namespace, device.DeviceID, device.IPAddress, device.Port)
	if err != nil {
		return &SyncResult{
			DeviceID: deviceID,
			Service:  serviceName,
			Status:   "failed",
			Error:    fmt.Sprintf("endpoints: %v", err),
		}, nil
	}

	status := "created"
	if wasExisting {
		status = "updated"
	}

	return &SyncResult{
		DeviceID: deviceID,
		Service:  serviceName,
		Status:   status,
	}, nil
}

// GetDeviceResources returns detailed K8s resource info for a device
func GetDeviceResources(namespace, deviceID string) (*DeviceResourceDetail, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))

	detail := &DeviceResourceDetail{
		DeviceID: deviceID,
		Service: ServiceInfo{
			Name:   serviceName,
			Exists: false,
			Ports:  []PortInfo{},
		},
		Endpoints: EndpointsInfo{
			Name:              serviceName,
			Exists:            false,
			ReadyAddresses:    []string{},
			NotReadyAddresses: []string{},
		},
		PrometheusTarget: fmt.Sprintf("http://%s.%s.svc:9100/metrics", serviceName, namespace),
	}

	// Get Service info
	service, err := GetService(namespace, deviceID)
	if err == nil && service != nil {
		detail.Service.Exists = true
		detail.Service.ClusterIP = service.Spec.ClusterIP
		for _, port := range service.Spec.Ports {
			detail.Service.Ports = append(detail.Service.Ports, PortInfo{
				Name: port.Name,
				Port: port.Port,
			})
		}
	}

	// Get Endpoints info
	endpoints, err := GetEndpoints(namespace, deviceID)
	if err == nil && endpoints != nil {
		detail.Endpoints.Exists = true
		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					detail.Endpoints.ReadyAddresses = append(detail.Endpoints.ReadyAddresses,
						fmt.Sprintf("%s:%d", addr.IP, port.Port))
				}
			}
			for _, addr := range subset.NotReadyAddresses {
				for _, port := range subset.Ports {
					detail.Endpoints.NotReadyAddresses = append(detail.Endpoints.NotReadyAddresses,
						fmt.Sprintf("%s:%d", addr.IP, port.Port))
				}
			}
		}
	}

	return detail, nil
}

// DeleteDeviceResources deletes K8s resources for a specific device
func DeleteDeviceResources(namespace, deviceID string) (*SyncResult, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))

	// Delete Service
	err := DeleteService(namespace, deviceID)
	if err != nil {
		return &SyncResult{
			DeviceID: deviceID,
			Service:  serviceName,
			Status:   "failed",
			Error:    fmt.Sprintf("delete service: %v", err),
		}, nil
	}

	// Delete Endpoints
	err = DeleteEndpoints(namespace, deviceID)
	if err != nil {
		return &SyncResult{
			DeviceID: deviceID,
			Service:  serviceName,
			Status:   "failed",
			Error:    fmt.Sprintf("delete endpoints: %v", err),
		}, nil
	}

	return &SyncResult{
		DeviceID: deviceID,
		Service:  serviceName,
		Status:   "deleted",
	}, nil
}

// getAllDevices fetches all devices from the server API
func getAllDevices(serverURL string) ([]models.DeviceStatus, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/devices", serverURL)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var listResponse models.DevicesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, err
	}

	return listResponse.Devices, nil
}

// getDeviceByID fetches a specific device from the server API
func getDeviceByID(serverURL, deviceID string) (*models.DeviceStatus, error) {
	devices, err := getAllDevices(serverURL)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.DeviceID == deviceID {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", deviceID)
}

// CleanupAllResources removes all edge-device-* resources from a namespace
func CleanupAllResources(namespace string) ([]string, []string, error) {
	if !IsInitialized() {
		return nil, nil, fmt.Errorf("kubernetes client not initialized")
	}

	// List all edge services
	services, err := ListEdgeServices(namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list services: %w", err)
	}

	// List all edge endpoints
	endpoints, err := ListEdgeEndpoints(namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Delete all services
	for _, serviceName := range services {
		deviceID := serviceName[len("edge-device-"):]
		if err := DeleteService(namespace, deviceID); err != nil {
			return nil, nil, fmt.Errorf("failed to delete service %s: %w", serviceName, err)
		}
	}

	// Delete all endpoints
	for _, endpointName := range endpoints {
		deviceID := endpointName[len("edge-device-"):]
		if err := DeleteEndpoints(namespace, deviceID); err != nil {
			return nil, nil, fmt.Errorf("failed to delete endpoints %s: %w", endpointName, err)
		}
	}

	return services, endpoints, nil
}
