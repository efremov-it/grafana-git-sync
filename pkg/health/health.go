package health

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status of the application
type Status struct {
	Status         string    `json:"status"` // "healthy", "degraded", "unhealthy"
	Timestamp      time.Time `json:"timestamp"`
	GrafanaHealthy bool      `json:"grafana_healthy"`
	GitSyncHealthy bool      `json:"git_sync_healthy"`
	LastSyncTime   time.Time `json:"last_sync_time,omitempty"`
	LastError      string    `json:"last_error,omitempty"`
}

// Checker manages health check state
type Checker struct {
	mu             sync.RWMutex
	grafanaHealthy bool
	gitSyncHealthy bool
	lastSyncTime   time.Time
	lastError      string
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		grafanaHealthy: false,
		gitSyncHealthy: false,
	}
}

// SetGrafanaHealth updates Grafana connectivity status
func (c *Checker) SetGrafanaHealth(healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.grafanaHealthy = healthy
}

// SetGitSyncHealth updates Git sync status
func (c *Checker) SetGitSyncHealth(healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gitSyncHealthy = healthy
}

// SetLastSync updates the last successful sync time
func (c *Checker) SetLastSync(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSyncTime = t
}

// SetLastError updates the last error message
func (c *Checker) SetLastError(err string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastError = err
}

// GetStatus returns current health status
func (c *Checker) GetStatus() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := "healthy"
	if !c.grafanaHealthy || !c.gitSyncHealthy {
		status = "degraded"
	}
	if !c.grafanaHealthy && !c.gitSyncHealthy {
		status = "unhealthy"
	}

	return Status{
		Status:         status,
		Timestamp:      time.Now(),
		GrafanaHealthy: c.grafanaHealthy,
		GitSyncHealthy: c.gitSyncHealthy,
		LastSyncTime:   c.lastSyncTime,
		LastError:      c.lastError,
	}
}

// Handler returns an HTTP handler for health checks
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := c.GetStatus()

		w.Header().Set("Content-Type", "application/json")

		// Set HTTP status code based on health
		switch status.Status {
		case "healthy":
			w.WriteHeader(http.StatusOK)
		case "degraded":
			w.WriteHeader(http.StatusOK) // Still return 200 for degraded
		case "unhealthy":
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			log.Printf("Failed to encode health status: %v", err)
		}
	}
}

// StartServer starts the health check HTTP server
func (c *Checker) StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", c.Handler())
	mux.HandleFunc("/health", c.Handler()) // Alternative endpoint
	mux.HandleFunc("/", c.Handler())       // Root endpoint

	log.Printf("Starting health check server on %s", addr)
	return http.ListenAndServe(addr, mux)
}
