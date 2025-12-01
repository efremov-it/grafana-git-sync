# Configuration

Complete reference for all environment variables and configuration options.

## Required Variables

### Git Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `GIT_REPO_URL` | Full Git repository URL (SSH or HTTPS) | `ssh://git@github.com/org/dashboards.git` or `https://github.com/org/dashboards.git` |
| `GIT_BRANCH` | Branch or tag to sync | `main`, `master`, `v1.0.0` |

### Grafana Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `GRAFANA_URL` | Grafana instance URL | `http://localhost:3000` or `https://grafana.example.com` |

## Authentication

### Git Authentication (Choose One)

**SSH Authentication:**
```bash
GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)"
```
- Private key with `\n` as newlines
- Used for `ssh://` or `git@` URLs

**HTTPS Authentication:**
```bash
GIT_HTTPS_USER=your-username
GIT_HTTPS_PASS=ghp_yourPersonalAccessToken
```
- Username + password or personal access token
- Used for `https://` URLs

### Grafana Authentication (Choose One)

**Admin Credentials (Auto-creates token):**
```bash
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=admin
```
- Creates service account token automatically
- Recommended for first-time setup

**Existing Service Account Token:**
```bash
GF_SECURITY_TOKEN=glsa_YourServiceAccountTokenHere
```
- Use pre-created Grafana service account token
- Recommended for production

## Optional Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GIT_LOCAL_REPO_DIR` | Local directory for Git clone | `/tmp/git-repo` | `/data/repo` |
| `GIT_REPO_SUBDIR` | Subdirectory containing dashboards | `.` (root) | `dashboards`, `grafana/dashboards` |
| `DASHBOARDS_DIR` | Temporary dashboard storage | `/tmp/dashboards` | `/data/dashboards` |
| `POLL_INTERVAL_SEC` | Git polling interval in seconds | `60` | `30`, `120` |
| `HEALTH_CHECK_PORT` | Health check HTTP server port | `8080` | `9090` |

## Configuration Examples

### Minimal Configuration (SSH)
```bash
GIT_REPO_URL=ssh://git@github.com/org/dashboards.git
GIT_BRANCH=main
GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)"
GRAFANA_URL=http://localhost:3000
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=admin
```

### Production Configuration (HTTPS + Token)
```bash
GIT_REPO_URL=https://github.com/org/dashboards.git
GIT_BRANCH=main
GIT_HTTPS_USER=bot-user
GIT_HTTPS_PASS=ghp_xxxxxxxxxxxxx
GIT_REPO_SUBDIR=grafana/dashboards
GRAFANA_URL=https://grafana.example.com
GF_SECURITY_TOKEN=glsa_xxxxxxxxxxxxx
POLL_INTERVAL_SEC=30
```

### Monorepo Configuration
```bash
GIT_REPO_URL=ssh://git@gitlab.com/company/monorepo.git
GIT_BRANCH=production
GIT_SSH_KEY="$(cat ~/.ssh/gitlab_key)"
GIT_REPO_SUBDIR=monitoring/grafana/dashboards  # Deep nesting supported
GRAFANA_URL=http://grafana:3000
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=secret
```

## Health Check Endpoint

The health check endpoint provides JSON status:

```bash
curl http://localhost:8080/healthz
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-01T03:45:00Z",
  "grafana_healthy": true,
  "git_sync_healthy": true,
  "last_sync_time": "2025-12-01T03:44:30Z",
  "last_error": ""
}
```

**Status Values:**
- `healthy` - Both Grafana and Git sync are working
- `degraded` - One service is down
- `unhealthy` - Both services are down (returns HTTP 503)

## Dashboard Versioning

When dashboards are uploaded, version metadata is automatically added:

**Version Message Format:**
```
commit abc1234: Updated CPU metrics - John Doe
```

This appears in Grafana's dashboard version history, linking each change to its Git commit.

## Environment Variable Priority

1. Environment variables (highest)
2. Default values (fallback)

No configuration files are used - everything is configured via environment variables for 12-factor app compliance.
