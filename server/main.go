package main

import (
	"context"
	"edge-metrics-server/database"
	"edge-metrics-server/handlers"
	"edge-metrics-server/kubernetes"
	"edge-metrics-server/models"
	"edge-metrics-server/repository"
	"edge-metrics-server/router"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./config.db"
	}

	syncInterval := 0 // 0 = disabled (Helm ServiceMonitor handles Prometheus targets)
	if v := os.Getenv("SYNC_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			syncInterval = n
		}
	}

	syncNamespace := os.Getenv("SYNC_NAMESPACE")
	if syncNamespace == "" {
		syncNamespace = "monitoring"
	}

	// Open database
	db, err := database.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create store (runs migrations on boot)
	store, err := repository.NewSQLiteStore(db)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Seed devices from environment (if provided)
	if seedJSON := os.Getenv("SEED_DEVICES"); seedJSON != "" {
		seedDevices(store, seedJSON)
	}

	// Inject store into handler
	h := handlers.NewHandler(store, store, store)

	// Initialize Kubernetes client (optional)
	var syncer *handlers.Syncer
	if err := kubernetes.InitClient(); err != nil {
		log.Printf("Kubernetes client not initialized: %v (Kubernetes features disabled)", err)
	} else {
		log.Printf("Kubernetes client initialized successfully")

		if syncInterval > 0 {
			// Start auto-sync goroutine (creates headless Service/Endpoints per device)
			syncer = handlers.NewSyncer(store, syncNamespace, time.Duration(syncInterval)*time.Second)
			h.SetSyncer(syncer)
			syncer.Start()
		} else {
			log.Printf("Auto-sync disabled (SYNC_INTERVAL=0)")
		}
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Setup routes with handler
	router.SetupRoutes(r, h)

	// Create http.Server for graceful shutdown
	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	// Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting CONFIG SERVER on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until signal
	sig := <-quit
	log.Printf("Received signal %v, shutting down...", sig)

	// Graceful shutdown with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Printf("Server shutdown gracefully")
	}

	// Stop syncer and wait for it to finish
	if syncer != nil {
		syncer.Stop()
	}

	log.Printf("Exiting")
}

func seedDevices(store *repository.SQLiteStore, seedJSON string) {
	var devices []struct {
		DeviceID   string `json:"device_id"`
		DeviceType string `json:"device_type"`
		IPAddress  string `json:"ip_address"`
	}
	if err := json.Unmarshal([]byte(seedJSON), &devices); err != nil {
		log.Printf("Warning: SEED_DEVICES parse error: %v", err)
		return
	}
	seeded := 0
	for _, d := range devices {
		exists, _ := store.Exists(d.DeviceID)
		if !exists {
			if err := store.Create(&models.DeviceConfig{
				DeviceID:   d.DeviceID,
				DeviceType: d.DeviceType,
				IPAddress:  d.IPAddress,
				Port:       9102,
				ReloadPort: 9101,
			}); err != nil {
				log.Printf("Warning: failed to seed %s: %v", d.DeviceID, err)
				continue
			}
			seeded++
		}
	}
	log.Printf("Seed complete: %d new devices (of %d total)", seeded, len(devices))
}
