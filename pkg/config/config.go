package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	RepoURL       string
	Branch        string
	SSHKey        string
	HTTPSUser     string
	HTTPSPassword string
	RepoDir       string
	RepoSubdir    string
	DashboardsDir string
	PollInterval  time.Duration
	GrafanaURL    string
	GrafanaUser   string
	GrafanaPass   string
	GrafanaToken  string
}

// Load reads and validates configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		RepoURL:       getEnvRequired("GIT_REPO_URL"),
		Branch:        getEnvRequired("GIT_BRANCH"),
		SSHKey:        os.Getenv("GIT_SSH_KEY"),
		HTTPSUser:     os.Getenv("GIT_HTTPS_USER"),
		HTTPSPassword: os.Getenv("GIT_HTTPS_PASS"),
		RepoDir:       getEnv("GIT_LOCAL_REPO_DIR", "/tmp/grafana_data"),
		RepoSubdir:    getEnv("GIT_REPO_SUBDIR", ""),
		DashboardsDir: getEnv("DASHBOARDS_DIR", "/tmp/grafana_data"),
		GrafanaURL:    getEnvRequired("GRAFANA_URL"),
		GrafanaUser:   getEnv("GF_SECURITY_ADMIN_USER", ""),
		GrafanaPass:   getEnv("GF_SECURITY_ADMIN_PASSWORD", ""),
		GrafanaToken:  getEnv("GF_SECURITY_TOKEN", ""),
	}

	pollIntervalStr := getEnv("POLL_INTERVAL_SEC", "60")
	pollIntervalSec, err := strconv.Atoi(pollIntervalStr)
	if err != nil || pollIntervalSec <= 0 {
		return nil, fmt.Errorf("invalid POLL_INTERVAL_SEC value: %s", pollIntervalStr)
	}
	cfg.PollInterval = time.Duration(pollIntervalSec) * time.Second

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that the configuration has valid authentication
func (c *Config) validate() error {
	// Check Git authentication
	hasSSH := c.SSHKey != ""
	hasHTTPS := c.HTTPSUser != "" && c.HTTPSPassword != ""
	if !hasSSH && !hasHTTPS {
		return fmt.Errorf("no Git authentication provided (SSH key or HTTPS credentials required)")
	}

	// Check Grafana authentication
	hasBasicAuth := c.GrafanaUser != "" && c.GrafanaPass != ""
	hasToken := c.GrafanaToken != ""
	if !hasBasicAuth && !hasToken {
		return fmt.Errorf("no Grafana authentication provided")
	}

	return nil
}

// SafeForLog returns a copy of the config with sensitive fields masked
func (c *Config) SafeForLog() *Config {
	masked := maskSensitiveFields(c).(Config)
	return &masked
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Required environment variable %s is not set", key)
	}
	return val
}

func maskSensitiveFields(input any) any {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	if t.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return input
	}

	masked := reflect.New(t).Elem()

	sensitivePatterns := []string{
		"password", "pass", "token", "secret",
		"key", "private", "credential", "ssh", "cert",
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		nameLower := strings.ToLower(field.Name)
		isSensitive := false
		for _, p := range sensitivePatterns {
			if strings.Contains(nameLower, p) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			if value.Kind() == reflect.String && value.String() != "" {
				masked.Field(i).SetString("***")
			} else {
				masked.Field(i).Set(value)
			}
		} else {
			masked.Field(i).Set(value)
		}
	}

	return masked.Interface()
}
