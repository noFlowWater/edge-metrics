package handlers

import (
	"edge-metrics-server/kubernetes"
	"edge-metrics-server/repository"
	"fmt"
	"log"
	"strings"
	gosync "sync"
	"time"
)

// Syncer handles periodic Kubernetes Service/Endpoints reconciliation.
type Syncer struct {
	mu            gosync.Mutex
	devices       repository.DeviceStore
	namespace     string
	interval      time.Duration
	resetCh       chan struct{}
	stopCh        chan struct{}
	done          chan struct{}
	failureCounts map[string]int
	graceLimit    int
}

// NewSyncer creates a Syncer with the given configuration.
func NewSyncer(devices repository.DeviceStore, namespace string, interval time.Duration) *Syncer {
	return &Syncer{
		devices:       devices,
		namespace:     namespace,
		interval:      interval,
		resetCh:       make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
		done:          make(chan struct{}),
		failureCounts: make(map[string]int),
		graceLimit:    3,
	}
}

// Start begins the periodic sync goroutine.
func (s *Syncer) Start() {
	log.Printf("auto-sync started, interval=%s", s.interval)
	go s.loop()
}

// Stop terminates the periodic sync goroutine and waits for it to finish.
func (s *Syncer) Stop() {
	close(s.stopCh)
	<-s.done
	log.Printf("auto-sync stopped")
}

func (s *Syncer) loop() {
	defer close(s.done)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			resp, err := s.Sync()
			if err != nil {
				log.Printf("auto-sync error: %v", err)
			} else {
				log.Printf("auto-sync result: created=%d updated=%d deleted=%d failed=%d",
					len(resp.Created), len(resp.Updated), len(resp.Deleted), len(resp.Failed))
			}
		case <-s.resetCh:
			ticker.Reset(s.interval)
		}
	}
}

// Sync performs one sync cycle. Safe for concurrent use.
func (s *Syncer) Sync() (*kubernetes.SyncResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !kubernetes.IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	allDevices, err := s.devices.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	response := &kubernetes.SyncResponse{
		Status:  "synced",
		Created: []kubernetes.SyncResult{},
		Updated: []kubernetes.SyncResult{},
		Deleted: []kubernetes.SyncResult{},
		Failed:  []kubernetes.SyncResult{},
	}

	existingServices, err := kubernetes.ListEdgeServices(s.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	existingMap := make(map[string]bool)
	for _, svc := range existingServices {
		existingMap[svc] = true
	}

	processed := make(map[string]bool)

	for _, device := range allDevices {
		serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(device.DeviceID))
		processed[serviceName] = true

		status := CheckDeviceHealth(device)
		if status.Status == "healthy" && device.IPAddress != "" {
			response.TotalHealthy++
			delete(s.failureCounts, device.DeviceID)

			wasExisting := existingMap[serviceName]

			if err := kubernetes.CreateOrUpdateService(s.namespace, device.DeviceID, device.DeviceType, device.Port); err != nil {
				response.Failed = append(response.Failed, kubernetes.SyncResult{
					DeviceID: device.DeviceID, Service: serviceName,
					Status: "failed", Error: fmt.Sprintf("service: %v", err),
				})
				continue
			}

			if err := kubernetes.CreateOrUpdateEndpoints(s.namespace, device.DeviceID, device.IPAddress, device.Port); err != nil {
				response.Failed = append(response.Failed, kubernetes.SyncResult{
					DeviceID: device.DeviceID, Service: serviceName,
					Status: "failed", Error: fmt.Sprintf("endpoints: %v", err),
				})
				continue
			}

			r := kubernetes.SyncResult{DeviceID: device.DeviceID, Service: serviceName}
			if wasExisting {
				r.Status = "updated"
				response.Updated = append(response.Updated, r)
			} else {
				r.Status = "created"
				response.Created = append(response.Created, r)
			}
		} else {
			s.failureCounts[device.DeviceID]++
			if s.failureCounts[device.DeviceID] >= s.graceLimit {
				// Grace period exceeded: remove Endpoints only, keep Service
				kubernetes.DeleteEndpoints(s.namespace, device.DeviceID)
				log.Printf("grace period exceeded for %s (%d failures): Endpoints removed",
					device.DeviceID, s.failureCounts[device.DeviceID])
			}
		}
	}

	// Clean up services for devices that no longer exist in the store
	for serviceName := range existingMap {
		if processed[serviceName] {
			continue
		}
		deviceID := serviceName[len("edge-device-"):]
		kubernetes.DeleteService(s.namespace, deviceID)
		kubernetes.DeleteEndpoints(s.namespace, deviceID)
		delete(s.failureCounts, deviceID)
		response.Deleted = append(response.Deleted, kubernetes.SyncResult{
			DeviceID: deviceID, Service: serviceName, Status: "deleted",
		})
	}

	return response, nil
}

// TriggerSync performs a sync and resets the periodic timer.
func (s *Syncer) TriggerSync() (*kubernetes.SyncResponse, error) {
	result, err := s.Sync()
	select {
	case s.resetCh <- struct{}{}:
	default:
	}
	return result, err
}
