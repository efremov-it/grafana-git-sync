package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"GRAFANA_URL":         "http://localhost:3000",
				"GF_SECURITY_TOKEN":   "test-token",
				"GIT_REPO_URL":        "https://github.com/test/repo.git",
				"GIT_BRANCH":          "main",
				"GIT_HTTPS_USER":      "user",
				"GIT_HTTPS_PASS":      "pass",
				"POLL_INTERVAL_SEC":   "30",
				"GIT_LOCAL_REPO_DIR":  "/tmp/dashboards",
			},
			wantErr: false,
		},
		{
			name: "missing required GRAFANA_URL",
			envVars: map[string]string{
				"GF_SECURITY_TOKEN":   "test-token",
				"GIT_REPO_URL":        "https://github.com/test/repo.git",
				"GIT_BRANCH":          "main",
				"GIT_HTTPS_USER":      "user",
				"GIT_HTTPS_PASS":      "pass",
				"POLL_INTERVAL_SEC":   "30",
				"GIT_LOCAL_REPO_DIR":  "/tmp/dashboards",
			},
			wantErr: true,
		},
		{
			name: "missing required GRAFANA_TOKEN",
			envVars: map[string]string{
			"GRAFANA_URL":         "http://localhost:3000",
			"GIT_REPO_URL":        "https://github.com/test/repo.git",
			"GIT_BRANCH":          "main",
			"POLL_INTERVAL_SEC":   "30",
			"GIT_LOCAL_REPO_DIR":  "/tmp/dashboards",
			},
			wantErr: true,
		},
		{
			name: "missing required GIT_REPO_URL",
			envVars: map[string]string{
				"GRAFANA_URL":      "http://localhost:3000",
				"GRAFANA_TOKEN":    "test-token",
				"GIT_BRANCH":       "main",
				"SYNC_INTERVAL":    "30",
				"DASHBOARD_FOLDER": "/tmp/dashboards",
			},
			wantErr: true,
		},
		{
			name: "default values applied",
			envVars: map[string]string{
				"GRAFANA_URL":       "http://localhost:3000",
				"GF_SECURITY_TOKEN": "test-token",
				"GIT_REPO_URL":      "https://github.com/test/repo.git",
				"GIT_BRANCH":        "main",
				"GIT_HTTPS_USER":    "user",
				"GIT_HTTPS_PASS":    "pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg.GrafanaURL != tt.envVars["GRAFANA_URL"] {
					t.Errorf("GrafanaURL = %v, want %v", cfg.GrafanaURL, tt.envVars["GRAFANA_URL"])
				}
				if cfg.GrafanaToken != tt.envVars["GF_SECURITY_TOKEN"] {
					t.Errorf("GrafanaToken = %v, want %v", cfg.GrafanaToken, tt.envVars["GF_SECURITY_TOKEN"])
				}
				if cfg.RepoURL != tt.envVars["GIT_REPO_URL"] {
					t.Errorf("RepoURL = %v, want %v", cfg.RepoURL, tt.envVars["GIT_REPO_URL"])
				}

				// Check defaults
				if tt.envVars["POLL_INTERVAL_SEC"] == "" && cfg.PollInterval != 60*time.Second {
					t.Errorf("PollInterval default = %v, want 60s", cfg.PollInterval)
				}
				if tt.envVars["GIT_LOCAL_REPO_DIR"] == "" && cfg.RepoDir != "/tmp/grafana_data" {
					t.Errorf("RepoDir default = %v, want '/tmp/grafana_data'", cfg.RepoDir)
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				GrafanaURL:    "http://localhost:3000",
				GrafanaToken:  "token",
				RepoURL:       "https://github.com/test/repo.git",
				Branch:        "main",
				PollInterval:  60 * time.Second,
				RepoDir:       "/tmp/dashboards",
			},
			wantErr: false,
		},
		{
			name: "missing GrafanaURL",
			config: &Config{
				GrafanaToken:  "token",
				RepoURL:       "https://github.com/test/repo.git",
				Branch:        "main",
				PollInterval:  60 * time.Second,
				RepoDir:       "/tmp/dashboards",
			},
			wantErr: true,
		},
		{
			name: "missing RepoURL",
			config: &Config{
				GrafanaURL:    "http://localhost:3000",
				GrafanaToken:  "token",
				Branch:        "main",
				PollInterval:  60 * time.Second,
				RepoDir:       "/tmp/dashboards",
			},
			wantErr: true,
		},
		{
			name: "missing Branch",
			config: &Config{
				GrafanaURL:    "http://localhost:3000",
				GrafanaToken:  "token",
				RepoURL:       "https://github.com/test/repo.git",
				PollInterval:  60 * time.Second,
				RepoDir:       "/tmp/dashboards",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_SafeForLog(t *testing.T) {
	cfg := &Config{
		GrafanaURL:    "http://localhost:3000",
		GrafanaToken:  "secret-token-12345",
		GrafanaPass:   "secret-password",
		RepoURL:       "https://user:password@github.com/test/repo.git",
		HTTPSUser:     "testuser",
		HTTPSPassword: "secret-password",
		Branch:        "main",
		PollInterval:  60 * time.Second,
		RepoDir:       "/tmp/dashboards",
	}

	safeCfg := cfg.SafeForLog()

	if safeCfg.GrafanaToken != "***" {
		t.Errorf("SafeForLog() did not mask GrafanaToken, got %v", safeCfg.GrafanaToken)
	}
	if safeCfg.GrafanaPass != "***" {
		t.Errorf("SafeForLog() did not mask GrafanaPass, got %v", safeCfg.GrafanaPass)
	}
	if safeCfg.HTTPSPassword != "***" {
		t.Errorf("SafeForLog() did not mask HTTPSPassword, got %v", safeCfg.HTTPSPassword)
	}
	if safeCfg.RepoURL == cfg.RepoURL {
		t.Errorf("SafeForLog() did not mask credentials in RepoURL")
	}
	if safeCfg.GrafanaURL != cfg.GrafanaURL {
		t.Errorf("SafeForLog() modified GrafanaURL, got %v", safeCfg.GrafanaURL)
	}
}
