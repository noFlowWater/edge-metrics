package handlers

import "edge-metrics-server/repository"

// Handler holds injected store dependencies for all HTTP handlers.
type Handler struct {
	devices repository.DeviceStore
	health  repository.HealthStore
	sync    repository.SyncStore
	syncer  *Syncer
}

// NewHandler creates a Handler with the required store interfaces.
func NewHandler(devices repository.DeviceStore, health repository.HealthStore, sync repository.SyncStore) *Handler {
	return &Handler{
		devices: devices,
		health:  health,
		sync:    sync,
	}
}

// SetSyncer attaches an auto-sync goroutine manager to the handler.
func (h *Handler) SetSyncer(s *Syncer) {
	h.syncer = s
}
