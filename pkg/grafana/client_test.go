package grafana

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:3000", "test-token", "", "")
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.url != "http://localhost:3000" {
		t.Errorf("url = %v, want %v", client.url, "http://localhost:3000")
	}
	if client.token != "test-token" {
		t.Errorf("token = %v, want %v", client.token, "test-token")
	}
	if client.folders == nil {
		t.Error("folders map not initialized")
	}
}

func TestClient_GetFolderByTitle(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		statusCode   int
		title        string
		parentUid    string
		expectID     int
		expectUID    string
		expectError  bool
	}{
		{
			name: "folder found",
			mockResponse: `[
				{"id": 1, "uid": "uid1", "title": "TestFolder", "parentUid": ""},
				{"id": 2, "uid": "uid2", "title": "TestFolder", "parentUid": "parent1"}
			]`,
			statusCode:  200,
			title:       "TestFolder",
			parentUid:   "parent1",
			expectID:    2,
			expectUID:   "uid2",
			expectError: false,
		},
		{
			name: "folder not found",
			mockResponse: `[
				{"id": 1, "uid": "uid1", "title": "OtherFolder", "parentUid": ""}
			]`,
			statusCode:  200,
			title:       "TestFolder",
			parentUid:   "",
			expectID:    0,
			expectUID:   "",
			expectError: false,
		},
		{
			name:         "API error",
			mockResponse: `{"message": "Internal server error"}`,
			statusCode:   500,
			title:        "TestFolder",
			parentUid:    "",
			expectID:     0,
			expectUID:    "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/folders" {
					t.Errorf("Expected path /api/folders, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", "", "")
			id, uid, err := client.getFolderByTitle(tt.title, tt.parentUid)

			if (err != nil) != tt.expectError {
				t.Errorf("getFolderByTitle() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if id != tt.expectID {
				t.Errorf("getFolderByTitle() id = %v, want %v", id, tt.expectID)
			}
			if uid != tt.expectUID {
				t.Errorf("getFolderByTitle() uid = %v, want %v", uid, tt.expectUID)
			}
		})
	}
}

func TestClient_UploadDashboard(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		mockResponse string
		expectError  bool
	}{
		{
			name:         "successful upload",
			statusCode:   200,
			mockResponse: `{"status": "success", "uid": "dash-uid"}`,
			expectError:  false,
		},
		{
			name:         "API error",
			statusCode:   400,
			mockResponse: `{"message": "Invalid dashboard"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/dashboards/db" {
					t.Errorf("Expected path /api/dashboards/db, got %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", "", "")
			dashboard := map[string]interface{}{
				"dashboard": map[string]interface{}{
					"title": "Test Dashboard",
				},
			}

			err := client.UploadDashboard(dashboard, 1)
			if (err != nil) != tt.expectError {
				t.Errorf("UploadDashboard() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestSplitFolderPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "folder1/folder2/folder3",
			expected: []string{"folder1", "folder2", "folder3"},
		},
		{
			name:     "path with leading slash",
			path:     "/folder1/folder2",
			expected: []string{"folder1", "folder2"},
		},
		{
			name:     "path with trailing slash",
			path:     "folder1/folder2/",
			expected: []string{"folder1", "folder2"},
		},
		{
			name:     "single folder",
			path:     "folder1",
			expected: []string{"folder1"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitFolderPath(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("splitFolderPath() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitFolderPath()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestClient_CreateFolder(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		mockResponse string
		expectError  bool
	}{
		{
			name:         "successful creation",
			statusCode:   200,
			mockResponse: `{"id": 1, "uid": "folder-uid", "title": "Test"}`,
			expectError:  false,
		},
		{
			name:         "folder already exists",
			statusCode:   409,
			mockResponse: `{"message": "Folder already exists"}`,
			expectError:  true,
		},
		{
			name:         "API error",
			statusCode:   500,
			mockResponse: `{"message": "Internal error"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/folders" {
					if r.Method == "GET" {
						// Return empty list for existence check
						w.WriteHeader(200)
						w.Write([]byte("[]"))
					} else if r.Method == "POST" {
						w.WriteHeader(tt.statusCode)
						w.Write([]byte(tt.mockResponse))
					}
				} else {
					w.WriteHeader(404)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", "", "")
			_, err := client.createFolderRecursive("test-folder", "")

			if (err != nil) != tt.expectError {
				t.Errorf("createFolderRecursive() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
