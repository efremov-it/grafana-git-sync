# Systemd Deployment Guide

This guide covers deploying Grafana Git Sync as a systemd service on Linux.

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Service Management](#service-management)
- [Troubleshooting](#troubleshooting)

---

## 🔧 Prerequisites

- Linux system with systemd
- Go 1.24+ (for building from source) or download binary
- Git installed
- Network access to Git repository and Grafana

---

## 📦 Installation

### Step 1: Create User and Directories

```bash
# Create service user (no login shell)
sudo useradd -r -s /bin/false grafana-git-sync

# Create directories
sudo mkdir -p /opt/grafana-git-sync
sudo mkdir -p /var/lib/grafana-git-sync/{repo,dashboards}
sudo mkdir -p /etc/grafana-git-sync

# Set permissions
sudo chown -R grafana-git-sync:grafana-git-sync /var/lib/grafana-git-sync
sudo chown -R grafana-git-sync:grafana-git-sync /opt/grafana-git-sync
```

### Step 2: Install Binary

**Option A: Download from GitHub Releases**

```bash
# Download latest release
VERSION="v0.1.0"
wget https://github.com/efremov-it/grafana-git-sync/releases/download/${VERSION}/grafana-git-sync-linux-amd64

# Make executable
chmod +x grafana-git-sync-linux-amd64

# Move to system path
sudo mv grafana-git-sync-linux-amd64 /usr/local/bin/grafana-git-sync
```

**Option B: Build from Source**

```bash
# Clone repository
git clone https://github.com/efremov-it/grafana-git-sync.git
cd grafana-git-sync

# Build
make build

# Install
sudo cp bin/grafana-git-sync /usr/local/bin/
sudo chmod +x /usr/local/bin/grafana-git-sync
```

### Step 3: Install Systemd Service

```bash
# Copy service file
sudo cp examples/grafana-git-sync.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload
```

---

## ⚙️ Configuration

### Configure Git Authentication

**For SSH Authentication:**

```bash
# Create SSH key (if needed)
sudo -u grafana-git-sync ssh-keygen -t ed25519 -f /opt/grafana-git-sync/.ssh/id_ed25519 -N ""

# Add public key to your Git provider (GitHub/GitLab/Bitbucket)
cat /opt/grafana-git-sync/.ssh/id_ed25519.pub

# Create environment file with SSH key
sudo bash -c 'cat > /etc/grafana-git-sync/ssh-key.env << EOF
GIT_SSH_KEY="$(cat /opt/grafana-git-sync/.ssh/id_ed25519)"
EOF'

# Secure the file
sudo chmod 600 /etc/grafana-git-sync/ssh-key.env
sudo chown grafana-git-sync:grafana-git-sync /etc/grafana-git-sync/ssh-key.env
```

**For HTTPS Authentication:**

```bash
# Create environment file
sudo bash -c 'cat > /etc/grafana-git-sync/git-https.env << EOF
GIT_HTTPS_USER=your-username
GIT_HTTPS_PASS=your-token-or-password
EOF'

# Secure the file
sudo chmod 600 /etc/grafana-git-sync/git-https.env
sudo chown grafana-git-sync:grafana-git-sync /etc/grafana-git-sync/git-https.env
```

### Configure Grafana Authentication

**For Auto-Created Token:**

```bash
# Create password file
sudo bash -c 'cat > /etc/grafana-git-sync/grafana-password.env << EOF
GF_SECURITY_ADMIN_PASSWORD=your-admin-password
EOF'

# Secure the file
sudo chmod 600 /etc/grafana-git-sync/grafana-password.env
sudo chown grafana-git-sync:grafana-git-sync /etc/grafana-git-sync/grafana-password.env
```

**For Existing Token:**

```bash
# Create token file
sudo bash -c 'cat > /etc/grafana-git-sync/grafana-token.env << EOF
GF_SECURITY_TOKEN=glsa_your_service_account_token
EOF'

# Secure the file
sudo chmod 600 /etc/grafana-git-sync/grafana-token.env
sudo chown grafana-git-sync:grafana-git-sync /etc/grafana-git-sync/grafana-token.env
```

### Edit Service Configuration

```bash
# Edit service file
sudo systemctl edit --full grafana-git-sync.service
```

**Required changes:**

1. Update `GIT_REPO_URL` with your repository
2. Update `GIT_BRANCH` with your branch name
3. Update `GRAFANA_URL` if not localhost:3000
4. Uncomment appropriate authentication method
5. Optional: Set `GIT_REPO_SUBDIR` for monorepos

**Example configuration:**

```ini
# Git Configuration
Environment="GIT_REPO_URL=ssh://git@github.com/myorg/dashboards.git"
Environment="GIT_BRANCH=main"
Environment="GIT_REPO_SUBDIR=grafana/production"

# Grafana Configuration
Environment="GRAFANA_URL=http://grafana.example.com:3000"
Environment="GF_SECURITY_ADMIN_USER=admin"
EnvironmentFile=-/etc/grafana-git-sync/grafana-password.env
```

---

## 🚀 Service Management

### Start Service

```bash
# Enable service (start on boot)
sudo systemctl enable grafana-git-sync

# Start service
sudo systemctl start grafana-git-sync

# Check status
sudo systemctl status grafana-git-sync
```

### Monitor Logs

```bash
# Follow logs
sudo journalctl -u grafana-git-sync -f

# View recent logs
sudo journalctl -u grafana-git-sync -n 100

# View logs since boot
sudo journalctl -u grafana-git-sync -b

# View logs for specific time
sudo journalctl -u grafana-git-sync --since "1 hour ago"
```

### Control Service

```bash
# Stop service
sudo systemctl stop grafana-git-sync

# Restart service
sudo systemctl restart grafana-git-sync

# Reload configuration
sudo systemctl daemon-reload
sudo systemctl restart grafana-git-sync

# Disable service (prevent auto-start)
sudo systemctl disable grafana-git-sync
```

### Health Check

```bash
# Check health endpoint
curl http://localhost:8080/healthz | jq

# Expected output:
# {
#   "status": "healthy",
#   "grafana_connected": true,
#   "git_connected": true,
#   "last_sync": "2024-12-01T10:30:00Z",
#   "sync_interval_sec": 60
# }
```

---

## 🔍 Troubleshooting

### Service Won't Start

**Check logs:**
```bash
sudo journalctl -u grafana-git-sync -n 50 --no-pager
```

**Common issues:**

1. **Permission denied on directories:**
   ```bash
   sudo chown -R grafana-git-sync:grafana-git-sync /var/lib/grafana-git-sync
   ```

2. **Binary not found:**
   ```bash
   ls -l /usr/local/bin/grafana-git-sync
   ```

3. **Environment file not readable:**
   ```bash
   sudo chmod 600 /etc/grafana-git-sync/*.env
   sudo chown grafana-git-sync:grafana-git-sync /etc/grafana-git-sync/*.env
   ```

### Git Authentication Fails

**SSH Key Issues:**

```bash
# Test SSH connection as service user
sudo -u grafana-git-sync ssh -T git@github.com

# Check SSH key permissions
ls -l /opt/grafana-git-sync/.ssh/
# Should be: -rw------- (600)

# Fix permissions if needed
sudo chmod 600 /opt/grafana-git-sync/.ssh/id_ed25519
sudo chown grafana-git-sync:grafana-git-sync /opt/grafana-git-sync/.ssh/id_ed25519
```

**HTTPS Authentication Issues:**

```bash
# Verify environment file
sudo cat /etc/grafana-git-sync/git-https.env

# Test Git access
sudo -u grafana-git-sync git ls-remote https://username:token@github.com/org/repo.git
```

### Grafana Connection Issues

**Check Grafana is accessible:**

```bash
curl http://localhost:3000/api/health

# From service user perspective
sudo -u grafana-git-sync curl http://localhost:3000/api/health
```

**Test Grafana credentials:**

```bash
curl -u admin:password http://localhost:3000/api/org
```

### Service Crashes or Restarts

**View crash information:**

```bash
sudo journalctl -u grafana-git-sync --since "10 minutes ago"
```

**Check system resources:**

```bash
# Memory usage
sudo systemctl status grafana-git-sync | grep Memory

# CPU usage
top -p $(pgrep grafana-git-sync)
```

**Adjust restart policy if needed:**

Edit service file:
```ini
[Service]
Restart=always
RestartSec=30s  # Increase restart delay
StartLimitBurst=5
StartLimitIntervalSec=600
```

---

## 🔐 Security Best Practices

### File Permissions

```bash
# Service binary
sudo chmod 755 /usr/local/bin/grafana-git-sync

# Configuration files
sudo chmod 600 /etc/grafana-git-sync/*.env

# Data directories
sudo chmod 750 /var/lib/grafana-git-sync
```

### SSH Key Management

```bash
# Use SSH agent forwarding (alternative to storing keys)
# Or use deploy keys with read-only access

# Rotate SSH keys regularly
sudo -u grafana-git-sync ssh-keygen -t ed25519 -f /opt/grafana-git-sync/.ssh/id_ed25519_new
# Update Git provider with new public key
# Update environment file
# Restart service
```

### Grafana Token Rotation

```bash
# Create new service account token in Grafana
# Update environment file
sudo bash -c 'cat > /etc/grafana-git-sync/grafana-token.env << EOF
GF_SECURITY_TOKEN=glsa_new_token_here
EOF'

# Restart service
sudo systemctl restart grafana-git-sync
```

---

## 📊 Monitoring

### Prometheus Integration (Future)

```bash
# Metrics endpoint (planned for future release)
curl http://localhost:8080/metrics
```

### Log Monitoring

**Send logs to external system:**

```bash
# Using syslog
sudo systemctl edit grafana-git-sync.service

# Add:
[Service]
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=grafana-git-sync
SyslogFacility=local0
```

### Alerting

**Monitor service status:**

```bash
# Create monitoring script
cat > /usr/local/bin/check-grafana-git-sync.sh << 'EOF'
#!/bin/bash
if ! systemctl is-active --quiet grafana-git-sync; then
    echo "grafana-git-sync is not running!"
    # Send alert (email, Slack, etc.)
    exit 1
fi

# Check health endpoint
if ! curl -sf http://localhost:8080/healthz > /dev/null; then
    echo "grafana-git-sync health check failed!"
    exit 1
fi

echo "grafana-git-sync is healthy"
exit 0
EOF

chmod +x /usr/local/bin/check-grafana-git-sync.sh

# Add to cron (every 5 minutes)
echo "*/5 * * * * /usr/local/bin/check-grafana-git-sync.sh" | sudo crontab -
```

---

## 🔄 Updating

### Update Binary

```bash
# Stop service
sudo systemctl stop grafana-git-sync

# Backup old binary
sudo cp /usr/local/bin/grafana-git-sync /usr/local/bin/grafana-git-sync.old

# Download new version
VERSION="v0.2.0"
wget https://github.com/efremov-it/grafana-git-sync/releases/download/${VERSION}/grafana-git-sync-linux-amd64
sudo mv grafana-git-sync-linux-amd64 /usr/local/bin/grafana-git-sync
sudo chmod +x /usr/local/bin/grafana-git-sync

# Start service
sudo systemctl start grafana-git-sync

# Check logs
sudo journalctl -u grafana-git-sync -f
```

### Rollback

```bash
# If update fails, rollback
sudo systemctl stop grafana-git-sync
sudo mv /usr/local/bin/grafana-git-sync.old /usr/local/bin/grafana-git-sync
sudo systemctl start grafana-git-sync
```

---

## 📖 Additional Resources

- [Configuration Guide](../configuration.md)
- [Architecture](../architecture.md)
- [Troubleshooting](../troubleshooting.md)
- [Docker Deployment](docker.md)
- [Kubernetes Deployment](kubernetes.md)

---

**Need help?** Open an issue on [GitHub](https://github.com/efremov-it/grafana-git-sync/issues)
