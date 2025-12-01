package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	checker := NewChecker()
	if checker == nil {
		t.Fatal("NewChecker returned nil")
	}

	status := checker.GetStatus()
	if status.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status initially, got %s", status.Status)
	}
}

func TestSetHealth(t *testing.T) {
	checker := NewChecker()

	checker.SetGrafanaHealth(true)
	checker.SetGitSyncHealth(true)

	status := checker.GetStatus()
	if status.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", status.Status)
	}
	if !status.GrafanaHealthy {
		t.Error("Expected GrafanaHealthy to be true")
	}
	if !status.GitSyncHealthy {
		t.Error("Expected GitSyncHealthy to be true")
	}
}

func TestDegradedStatus(t *testing.T) {
	checker := NewChecker()

	checker.SetGrafanaHealth(true)
	checker.SetGitSyncHealth(false)

	status := checker.GetStatus()
	if status.Status != "degraded" {
		t.Errorf("Expected degraded status, got %s", status.Status)
	}
}

func TestLastSync(t *testing.T) {
	checker := NewChecker()
	now := time.Now()

	checker.SetLastSync(now)

	status := checker.GetStatus()
	if !status.LastSyncTime.Equal(now) {
		t.Errorf("Expected LastSyncTime to be %v, got %v", now, status.LastSyncTime)
	}
}

func TestLastError(t *testing.T) {
	checker := NewChecker()
	errMsg := "test error"

	checker.SetLastError(errMsg)

	status := checker.GetStatus()
	if status.LastError != errMsg {
		t.Errorf("Expected LastError to be %s, got %s", errMsg, status.LastError)
	}
}

func TestHandler(t *testing.T) {
	checker := NewChecker()
	checker.SetGrafanaHealth(true)
	checker.SetGitSyncHealth(true)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler := checker.Handler()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var status Status
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Status != "healthy" {
		t.Errorf("Expected healthy status in response, got %s", status.Status)
	}
}

func TestHandlerUnhealthy(t *testing.T) {
	checker := NewChecker()
	// Both false = unhealthy

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler := checker.Handler()
	handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503, got %d", w.Code)
	}

	var status Status
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status in response, got %s", status.Status)
	}
}
