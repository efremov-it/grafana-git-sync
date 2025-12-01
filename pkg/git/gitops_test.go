package git

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		repoURL   string
		branch    string
		repoDir   string
		wantErr   bool
	}{
		{
			name:      "valid parameters with HTTPS",
			repoURL:   "https://github.com/test/repo.git",
			branch:    "main",
			repoDir:   tmpDir,
			wantErr:   false,
		},
		{
			name:      "valid parameters with SSH",
			repoURL:   "git@github.com:test/repo.git",
			branch:    "main",
			repoDir:   tmpDir,
			wantErr:   true, // SSH requires key
		},
		{
			name:      "empty repo URL",
			repoURL:   "",
			branch:    "main",
			repoDir:   tmpDir,
			wantErr:   true, // No auth provided
		},
	}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Provide dummy credentials for HTTPS URLs to pass auth check
				httpsUser, httpsPass := "", ""
				if tt.repoURL != "" && !tt.wantErr && (len(tt.repoURL) < 4 || tt.repoURL[:3] == "htt") {
					httpsUser, httpsPass = "user", "pass"
				}
				client, err := NewClient(tt.repoURL, tt.branch, tt.repoDir, "", httpsUser, httpsPass)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil without error")
			}
		})
	}
}

func TestClient_Clone_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	client, err := NewClient("invalid-url", "main", tmpDir, "", "user", "pass")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Clone()
	if err == nil {
		t.Error("Clone() should fail with invalid URL")
	}
}

func TestClient_FetchLatestCommit_NoRepo(t *testing.T) {
	tmpDir := t.TempDir()
	client, err := NewClient("https://github.com/test/repo.git", "main", tmpDir, "", "user", "pass")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Try to fetch without cloning first
	_, err = client.FetchLatestCommit()
	if err == nil {
		t.Error("FetchLatestCommit() should fail when repository is not cloned")
	}
}

func TestExtractCredentials(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		httpsUser     string
		httpsPass     string
		expectHasAuth bool
		wantErr       bool
	}{
		{
			name:          "URL without credentials, explicit auth",
			url:           "https://github.com/test/repo.git",
			httpsUser:     "myuser",
			httpsPass:     "mypass",
			expectHasAuth: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			client, err := NewClient(tt.url, "main", tmpDir, "", tt.httpsUser, tt.httpsPass)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				hasAuth := client.auth != nil
				if hasAuth != tt.expectHasAuth {
					t.Errorf("auth is nil = %v, want %v", !hasAuth, !tt.expectHasAuth)
				}
			}
		})
	}
}
