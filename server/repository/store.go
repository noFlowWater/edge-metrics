package repository

import "edge-metrics-server/models"

// DeviceStore provides CRUD operations for device configurations.
type DeviceStore interface {
	Get(deviceID string) (*models.DeviceConfig, error)
	List() ([]models.DeviceConfig, error)
	Create(config *models.DeviceConfig) error
	Update(deviceID string, config *models.DeviceConfig) error
	Upsert(deviceID string, config *models.DeviceConfig) (bool, error)
	Delete(deviceID string) error
	Exists(deviceID string) (bool, error)
}

// HealthStore tracks device health status in persistent storage.
type HealthStore interface {
	GetStatus(deviceID string) (string, error)
	ListStatuses() (map[string]string, error)
	UpdateHealth(deviceID, status string) error
}

// SyncStore tracks Kubernetes sync state per device.
type SyncStore interface {
	GetSyncState(deviceID string) (bool, error)
	MarkSynced(deviceID string) error
	ListSyncTargets() ([]string, error)
}
