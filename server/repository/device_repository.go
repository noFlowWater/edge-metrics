package repository

import (
	"database/sql"
	"edge-metrics-server/models"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// SQLiteStore implements DeviceStore, HealthStore, and SyncStore using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens the given db and runs migrations.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// --- migrations ---

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, `CREATE TABLE IF NOT EXISTS devices (
			device_id TEXT PRIMARY KEY,
			device_type TEXT NOT NULL,
			port INTEGER DEFAULT 9100,
			reload_port INTEGER DEFAULT 9101,
			enabled_metrics TEXT,
			extra_config TEXT,
			ip_address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`},
	}

	for _, m := range migrations {
		var count int
		s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&count)
		if count > 0 {
			continue
		}
		if _, err := s.db.Exec(m.sql); err != nil {
			return err
		}
		if _, err := s.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			return err
		}
		log.Printf("Applied migration %d", m.version)
	}

	// Legacy compatibility: add ip_address column if migrating from old schema
	s.db.Exec("ALTER TABLE devices ADD COLUMN ip_address TEXT")

	return nil
}

// --- DeviceStore implementation ---

func (s *SQLiteStore) Get(deviceID string) (*models.DeviceConfig, error) {
	query := `
		SELECT device_id, device_type, port, reload_port,
		       enabled_metrics, extra_config, ip_address
		FROM devices
		WHERE device_id = ?
	`

	var config models.DeviceConfig
	var enabledMetrics, extraConfig, ipAddress sql.NullString

	err := s.db.QueryRow(query, deviceID).Scan(
		&config.DeviceID,
		&config.DeviceType,
		&config.Port,
		&config.ReloadPort,
		&enabledMetrics,
		&extraConfig,
		&ipAddress,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if ipAddress.Valid {
		config.IPAddress = ipAddress.String
	}

	if enabledMetrics.Valid && enabledMetrics.String != "" {
		if err := json.Unmarshal([]byte(enabledMetrics.String), &config.EnabledMetrics); err != nil {
			return nil, err
		}
	}

	if extraConfig.Valid && extraConfig.String != "" {
		config.ExtraConfig = make(map[string]interface{})
		if err := json.Unmarshal([]byte(extraConfig.String), &config.ExtraConfig); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func (s *SQLiteStore) List() ([]models.DeviceConfig, error) {
	query := `
		SELECT device_id, device_type, port, reload_port,
		       enabled_metrics, extra_config, ip_address
		FROM devices
		ORDER BY device_id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.DeviceConfig
	for rows.Next() {
		var config models.DeviceConfig
		var enabledMetrics, extraConfig, ipAddress sql.NullString

		err := rows.Scan(
			&config.DeviceID,
			&config.DeviceType,
			&config.Port,
			&config.ReloadPort,
			&enabledMetrics,
			&extraConfig,
			&ipAddress,
		)
		if err != nil {
			return nil, err
		}

		if ipAddress.Valid {
			config.IPAddress = ipAddress.String
		}

		if enabledMetrics.Valid && enabledMetrics.String != "" {
			if err := json.Unmarshal([]byte(enabledMetrics.String), &config.EnabledMetrics); err != nil {
				return nil, fmt.Errorf("corrupt enabled_metrics for %s: %w", config.DeviceID, err)
			}
		}

		if extraConfig.Valid && extraConfig.String != "" {
			config.ExtraConfig = make(map[string]interface{})
			if err := json.Unmarshal([]byte(extraConfig.String), &config.ExtraConfig); err != nil {
				return nil, fmt.Errorf("corrupt extra_config for %s: %w", config.DeviceID, err)
			}
		}

		devices = append(devices, config)
	}

	return devices, rows.Err()
}

func (s *SQLiteStore) Create(config *models.DeviceConfig) error {
	var enabledMetrics, extraConfig sql.NullString

	if len(config.EnabledMetrics) > 0 {
		data, err := json.Marshal(config.EnabledMetrics)
		if err != nil {
			return err
		}
		enabledMetrics = sql.NullString{String: string(data), Valid: true}
	}

	if len(config.ExtraConfig) > 0 {
		data, err := json.Marshal(config.ExtraConfig)
		if err != nil {
			return err
		}
		extraConfig = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		INSERT INTO devices (device_id, device_type, port, reload_port,
		                    enabled_metrics, extra_config, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		config.DeviceID,
		config.DeviceType,
		config.Port,
		config.ReloadPort,
		enabledMetrics,
		extraConfig,
		config.IPAddress,
	)

	return err
}

func (s *SQLiteStore) Update(deviceID string, config *models.DeviceConfig) error {
	existing, err := s.Get(deviceID)
	if err != nil {
		return err
	}
	if existing == nil {
		return sql.ErrNoRows
	}

	var enabledMetrics, extraConfig sql.NullString

	if len(config.EnabledMetrics) > 0 {
		data, err := json.Marshal(config.EnabledMetrics)
		if err != nil {
			return err
		}
		enabledMetrics = sql.NullString{String: string(data), Valid: true}
	}

	if len(config.ExtraConfig) > 0 {
		data, err := json.Marshal(config.ExtraConfig)
		if err != nil {
			return err
		}
		extraConfig = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		UPDATE devices
		SET device_type = ?, port = ?, reload_port = ?,
		    enabled_metrics = ?, extra_config = ?, ip_address = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err = s.db.Exec(query,
		config.DeviceType,
		config.Port,
		config.ReloadPort,
		enabledMetrics,
		extraConfig,
		config.IPAddress,
		time.Now(),
		deviceID,
	)

	return err
}

func (s *SQLiteStore) Upsert(deviceID string, config *models.DeviceConfig) (bool, error) {
	exists, err := s.Exists(deviceID)
	if err != nil {
		return false, err
	}

	if exists {
		err = s.Update(deviceID, config)
		return false, err // false = updated
	}

	config.DeviceID = deviceID
	err = s.Create(config)
	return true, err // true = created
}

func (s *SQLiteStore) Delete(deviceID string) error {
	result, err := s.db.Exec("DELETE FROM devices WHERE device_id = ?", deviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *SQLiteStore) Exists(deviceID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE device_id = ?", deviceID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- HealthStore implementation (stub — backed by HTTP in handlers/health.go) ---

func (s *SQLiteStore) GetStatus(deviceID string) (string, error) {
	return "", nil
}

func (s *SQLiteStore) ListStatuses() (map[string]string, error) {
	return map[string]string{}, nil
}

func (s *SQLiteStore) UpdateHealth(deviceID, status string) error {
	return nil
}

// --- SyncStore implementation (stub — will be backed by DB in Phase 0.2) ---

func (s *SQLiteStore) GetSyncState(deviceID string) (bool, error) {
	return false, nil
}

func (s *SQLiteStore) MarkSynced(deviceID string) error {
	return nil
}

func (s *SQLiteStore) ListSyncTargets() ([]string, error) {
	devices, err := s.List()
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(devices))
	for i, d := range devices {
		ids[i] = d.DeviceID
	}
	return ids, nil
}
