package repository

import (
	"database/sql"
	"edge-metrics-server/models"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return store
}

func TestUpsertAndGet(t *testing.T) {
	s := newTestStore(t)
	cfg := &models.DeviceConfig{
		DeviceID:   "test-01",
		DeviceType: "jetson_orin",
		Port:       9100,
		ReloadPort: 9101,
		IPAddress:  "192.168.1.10",
	}
	created, err := s.Upsert("test-01", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("expected created=true for new device")
	}

	got, err := s.Get("test-01")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected device, got nil")
	}
	if got.DeviceType != "jetson_orin" {
		t.Errorf("expected device_type=jetson_orin, got %s", got.DeviceType)
	}
	if got.IPAddress != "192.168.1.10" {
		t.Errorf("expected ip=192.168.1.10, got %s", got.IPAddress)
	}
}

func TestListEmpty(t *testing.T) {
	s := newTestStore(t)
	devices, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestListMultiple(t *testing.T) {
	s := newTestStore(t)
	for _, id := range []string{"a", "b", "c"} {
		s.Create(&models.DeviceConfig{DeviceID: id, DeviceType: "stub", Port: 9100, ReloadPort: 9101, IPAddress: "1.2.3.4"})
	}
	devices, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}
}

func TestDeleteExisting(t *testing.T) {
	s := newTestStore(t)
	s.Create(&models.DeviceConfig{DeviceID: "del-me", DeviceType: "stub", Port: 9100, ReloadPort: 9101, IPAddress: "1.2.3.4"})
	err := s.Delete("del-me")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("del-me")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestDeleteNonExisting(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete("no-such-device")
	if err == nil {
		t.Error("expected error when deleting non-existing device")
	}
}

func TestUpdate(t *testing.T) {
	s := newTestStore(t)
	s.Create(&models.DeviceConfig{DeviceID: "upd", DeviceType: "stub", Port: 9100, ReloadPort: 9101, IPAddress: "1.2.3.4"})
	err := s.Update("upd", &models.DeviceConfig{DeviceType: "jetson_orin", Port: 9200, ReloadPort: 9201, IPAddress: "10.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("upd")
	if got.DeviceType != "jetson_orin" {
		t.Errorf("expected jetson_orin, got %s", got.DeviceType)
	}
	if got.Port != 9200 {
		t.Errorf("expected port 9200, got %d", got.Port)
	}
}

func TestMigrationIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s1, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatal(err)
	}
	_ = s1

	// Run migration again on same DB — should not fail
	s2, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatal(err)
	}
	_ = s2
}

func TestExists(t *testing.T) {
	s := newTestStore(t)
	exists, err := s.Exists("nope")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected false for non-existing device")
	}
	s.Create(&models.DeviceConfig{DeviceID: "yes", DeviceType: "stub", Port: 9100, ReloadPort: 9101, IPAddress: "1.2.3.4"})
	exists, err = s.Exists("yes")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected true for existing device")
	}
}

func TestGetNonExisting(t *testing.T) {
	s := newTestStore(t)
	got, err := s.Get("missing")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("expected nil for non-existing device")
	}
}

func TestEnabledMetricsRoundtrip(t *testing.T) {
	s := newTestStore(t)
	cfg := &models.DeviceConfig{
		DeviceID:       "metric-test",
		DeviceType:     "stub",
		Port:           9100,
		ReloadPort:     9101,
		IPAddress:      "1.2.3.4",
		EnabledMetrics: []string{"cpu_temp", "gpu_power"},
	}
	s.Create(cfg)
	got, _ := s.Get("metric-test")
	if len(got.EnabledMetrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(got.EnabledMetrics))
	}
}

func TestExtraConfigRoundtrip(t *testing.T) {
	s := newTestStore(t)
	cfg := &models.DeviceConfig{
		DeviceID:    "extra-test",
		DeviceType:  "stub",
		Port:        9100,
		ReloadPort:  9101,
		IPAddress:   "1.2.3.4",
		ExtraConfig: map[string]interface{}{"shelly": map[string]interface{}{"ip": "10.0.0.5"}},
	}
	s.Create(cfg)
	got, _ := s.Get("extra-test")
	if got.ExtraConfig == nil {
		t.Fatal("expected extra config, got nil")
	}
	if _, ok := got.ExtraConfig["shelly"]; !ok {
		t.Error("expected shelly key in extra config")
	}
}
