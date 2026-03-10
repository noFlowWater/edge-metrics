package handlers

import (
	"bytes"
	"database/sql"
	"edge-metrics-server/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockStore implements DeviceStore, HealthStore, SyncStore in-memory.
type mockStore struct {
	devices map[string]*models.DeviceConfig
}

func newMockStore() *mockStore {
	return &mockStore{devices: make(map[string]*models.DeviceConfig)}
}

func (m *mockStore) Get(id string) (*models.DeviceConfig, error) {
	d, ok := m.devices[id]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func (m *mockStore) List() ([]models.DeviceConfig, error) {
	var list []models.DeviceConfig
	for _, d := range m.devices {
		list = append(list, *d)
	}
	return list, nil
}

func (m *mockStore) Create(c *models.DeviceConfig) error {
	m.devices[c.DeviceID] = c
	return nil
}

func (m *mockStore) Update(id string, c *models.DeviceConfig) error {
	if _, ok := m.devices[id]; !ok {
		return sql.ErrNoRows
	}
	c.DeviceID = id
	m.devices[id] = c
	return nil
}

func (m *mockStore) Upsert(id string, c *models.DeviceConfig) (bool, error) {
	_, exists := m.devices[id]
	c.DeviceID = id
	m.devices[id] = c
	return !exists, nil
}

func (m *mockStore) Delete(id string) error {
	if _, ok := m.devices[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.devices, id)
	return nil
}

func (m *mockStore) Exists(id string) (bool, error) {
	_, ok := m.devices[id]
	return ok, nil
}

func (m *mockStore) GetStatus(id string) (string, error)              { return "", nil }
func (m *mockStore) ListStatuses() (map[string]string, error)         { return nil, nil }
func (m *mockStore) UpdateHealth(id, status string) error             { return nil }
func (m *mockStore) GetSyncState(id string) (bool, error)             { return false, nil }
func (m *mockStore) MarkSynced(id string) error                       { return nil }
func (m *mockStore) ListSyncTargets() ([]string, error)               { return nil, nil }

func setupRouter(store *mockStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(store, store, store)
	r := gin.New()
	r.GET("/health", h.Health)
	r.GET("/config", h.ListConfigs)
	r.GET("/config/:device_id", h.GetConfig)
	r.POST("/config/:device_id", h.CreateConfig)
	r.DELETE("/config/:device_id", h.DeleteConfig)
	r.GET("/devices", h.ListDevices)
	return r
}

func TestHealth(t *testing.T) {
	r := setupRouter(newMockStore())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateAndGet(t *testing.T) {
	store := newMockStore()
	r := setupRouter(store)

	body := `{"device_type":"jetson_orin","ip_address":"10.0.0.1","port":9100,"reload_port":9101}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/config/dev-01", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/config/dev-01", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["device_type"] != "jetson_orin" {
		t.Errorf("expected jetson_orin, got %v", resp["device_type"])
	}
}

func TestCreateDuplicate(t *testing.T) {
	store := newMockStore()
	r := setupRouter(store)

	body := `{"device_type":"stub","ip_address":"1.2.3.4"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/config/dup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("first create expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/config/dup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 409 {
		t.Errorf("duplicate expected 409, got %d", w.Code)
	}
}

func TestDelete(t *testing.T) {
	store := newMockStore()
	r := setupRouter(store)

	body := `{"device_type":"stub","ip_address":"1.2.3.4"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/config/del-me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/config/del-me", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	r := setupRouter(newMockStore())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/config/nope", nil)
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListConfigs(t *testing.T) {
	store := newMockStore()
	r := setupRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/config", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"].(float64) != 0 {
		t.Errorf("expected 0 total, got %v", resp["total"])
	}
}

func TestListDevices(t *testing.T) {
	store := newMockStore()
	r := setupRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
