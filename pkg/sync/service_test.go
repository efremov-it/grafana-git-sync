package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService("/tmp/repo", "subdir", "/tmp/dashboards")
	if service == nil {
		t.Error("NewService() returned nil")
	}
}

func TestLoadDashboard(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(tmpDir, "", tmpDir)

	// Create a valid dashboard file
	validDash := `{
		"dashboard": {
			"title": "Test Dashboard",
			"tags": ["test"]
		},
		"folderUid": "test-folder"
	}`

	validFile := filepath.Join(tmpDir, "valid.json")
	if err := os.WriteFile(validFile, []byte(validDash), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test valid dashboard
	dashboard, err := service.LoadDashboard(validFile)
	if err != nil {
		t.Errorf("LoadDashboard() error = %v", err)
	}
	if dashboard == nil {
		t.Error("LoadDashboard() returned nil dashboard")
	}

	// Test non-existent file
	_, err = service.LoadDashboard(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("LoadDashboard() should error on non-existent file")
	}

	// Test invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	_, err = service.LoadDashboard(invalidFile)
	if err == nil {
		t.Error("LoadDashboard() should error on invalid JSON")
	}
}

func TestCopyDashboards(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source structure
	folders := []string{
		"folder1",
		"folder1/subfolder1",
		"folder2",
	}

	for _, folder := range folders {
		path := filepath.Join(srcDir, folder)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		// Create dashboard files
		dashFile := filepath.Join(path, "dashboard.json")
		if err := os.WriteFile(dashFile, []byte(`{"dashboard": {}}`), 0644); err != nil {
			t.Fatalf("Failed to create test dashboard file: %v", err)
		}
	}

	service := NewService(srcDir, "", dstDir)
	files, err := service.CopyDashboards()
	if err != nil {
		t.Fatalf("CopyDashboards() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("CopyDashboards() returned no files")
	}

	// Verify destination structure
	for _, folder := range folders {
		path := filepath.Join(dstDir, folder)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Folder %s was not copied", folder)
		}

		dashFile := filepath.Join(path, "dashboard.json")
		if _, err := os.Stat(dashFile); os.IsNotExist(err) {
			t.Errorf("Dashboard file in %s was not copied", folder)
		}
	}
}

func TestCopyDashboards_WithSubdir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	subdir := "dashboards"

	// Create structure with subdirectory
	subdirPath := filepath.Join(srcDir, subdir)
	if err := os.MkdirAll(subdirPath, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create dashboard in subdirectory
	dashFile := filepath.Join(subdirPath, "test.json")
	if err := os.WriteFile(dashFile, []byte(`{"dashboard": {}}`), 0644); err != nil {
		t.Fatalf("Failed to create dashboard: %v", err)
	}

	service := NewService(srcDir, subdir, dstDir)
	files, err := service.CopyDashboards()
	if err != nil {
		t.Fatalf("CopyDashboards() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("CopyDashboards() returned no files")
	}
}
