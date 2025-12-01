# Grafana Git Sync - Examples

This directory contains working examples for deploying Grafana Git Sync in various environments.

## 📁 Contents

### Docker Compose

- **[docker-compose-ssh.yml](docker-compose-ssh.yml)** - Full stack with SSH authentication
- **[docker-compose-https.yml](docker-compose-https.yml)** - Full stack with HTTPS authentication
- **[docker-compose-token.yml](docker-compose-token.yml)** - Using existing Grafana service account token

### Kubernetes

- **[kubernetes.yaml](kubernetes.yaml)** - Complete K8s deployment with secrets, configmaps, deployment, and service

### Systemd

- **[grafana-git-sync.service](grafana-git-sync.service)** - Systemd service unit file for Linux

---

## 🚀 Quick Start

### Docker Compose

1. **Choose authentication method:**
   - SSH: `docker-compose-ssh.yml`
   - HTTPS: `docker-compose-https.yml`
   - Existing token: `docker-compose-token.yml`

2. **Edit the file:**
   ```bash
   # SSH: Add your private key to GIT_SSH_KEY
   # HTTPS: Set GIT_HTTPS_USER and GIT_HTTPS_PASS
   # Token: Set GF_SECURITY_TOKEN
   
   # Always set:
   GIT_REPO_URL=your-repo-url
   GIT_BRANCH=main
   ```

3. **Run:**
   ```bash
   docker-compose -f docker-compose-ssh.yml up -d
   ```

4. **Check health:**
   ```bash
   curl http://localhost:8080/healthz
   ```

5. **Access Grafana:**
   - URL: http://localhost:3000
   - Username: admin
   - Password: admin

---

### Systemd (Linux Service)

1. **Install binary:**
   ```bash
   # Download from GitHub releases or build from source
   sudo cp bin/grafana-git-sync /usr/local/bin/
   sudo chmod +x /usr/local/bin/grafana-git-sync
   ```

2. **Copy service file:**
   ```bash
   sudo cp grafana-git-sync.service /etc/systemd/system/
   sudo systemctl daemon-reload
   ```

3. **Configure authentication:**
   ```bash
   # Create configuration directory
   sudo mkdir -p /etc/grafana-git-sync
   
   # For SSH: Create SSH key environment file
   sudo bash -c 'cat > /etc/grafana-git-sync/ssh-key.env << EOF
   GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)"
   EOF'
   
   # For Grafana: Create password file
   sudo bash -c 'cat > /etc/grafana-git-sync/grafana-password.env << EOF
   GF_SECURITY_ADMIN_PASSWORD=your-password
   EOF'
   
   # Secure files
   sudo chmod 600 /etc/grafana-git-sync/*.env
   ```

4. **Edit service configuration:**
   ```bash
   sudo systemctl edit --full grafana-git-sync.service
   # Update GIT_REPO_URL, GIT_BRANCH, GRAFANA_URL
   ```

5. **Start service:**
   ```bash
   sudo systemctl enable --now grafana-git-sync
   sudo systemctl status grafana-git-sync
   sudo journalctl -u grafana-git-sync -f
   ```

**Full guide:** [docs/deployment/systemd.md](../docs/deployment/systemd.md)

---

### Kubernetes

1. **Edit secrets in `kubernetes.yaml`:**
   ```yaml
   # SSH key secret
   stringData:
     ssh-key: |
       -----BEGIN OPENSSH PRIVATE KEY-----
       ... your key here ...
       -----END OPENSSH PRIVATE KEY-----
   
   # Admin credentials secret
   stringData:
     username: admin
     password: your-strong-password
   ```

2. **Update ConfigMap:**
   ```yaml
   data:
     GIT_REPO_URL: "ssh://git@github.com/your-org/dashboards.git"
     GIT_BRANCH: "main"
     GRAFANA_URL: "http://grafana:3000"
   ```

3. **Apply:**
   ```bash
   kubectl apply -f kubernetes.yaml
   ```

4. **Check status:**
   ```bash
   kubectl get pods -n monitoring
   kubectl logs -n monitoring deployment/grafana-git-sync -f
   ```

5. **Test health endpoint:**
   ```bash
   kubectl port-forward -n monitoring svc/grafana-git-sync 8080:8080
   curl http://localhost:8080/healthz
   ```

---

## 🔧 Configuration Tips

### SSH Key Format

SSH keys must be in OpenSSH format with literal newlines:

```yaml
GIT_SSH_KEY: |
  -----BEGIN OPENSSH PRIVATE KEY-----
  b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABlwAAAAdzc2gtcn
  ... (multiple lines) ...
  -----END OPENSSH PRIVATE KEY-----
```

**Convert from PEM format:**
```bash
ssh-keygen -p -m OpenSSH -f ~/.ssh/id_rsa
```

### HTTPS Authentication

**GitHub:**
```yaml
GIT_HTTPS_USER: your-username
GIT_HTTPS_PASS: ghp_yourPersonalAccessToken
```

**GitLab:**
```yaml
GIT_HTTPS_USER: oauth2
GIT_HTTPS_PASS: glpat-yourAccessToken
```

### Monorepo Support

If dashboards are in a subdirectory:

```yaml
GIT_REPO_SUBDIR: dashboards/production
```

Repository structure:
```
repo/
├── app/
├── dashboards/
│   └── production/  ← Set GIT_REPO_SUBDIR to this
│       ├── folder1/
│       └── folder2/
└── docs/
```

---

## 📊 Health Check Examples

### JSON Response

```bash
curl http://localhost:8080/healthz | jq
```

```json
{
  "status": "healthy",
  "grafana_connected": true,
  "git_connected": true,
  "last_sync": "2024-01-15T10:30:00Z",
  "sync_interval_sec": 60
}
```

### Use in Docker Compose

```yaml
healthcheck:
  test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

### Use in Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

---

## 🔍 Troubleshooting

### Check logs

**Docker:**
```bash
docker logs grafana-git-sync -f
```

**Kubernetes:**
```bash
kubectl logs -n monitoring deployment/grafana-git-sync -f
```

### Test Git connectivity

**Docker:**
```bash
docker exec grafana-git-sync sh -c "git ls-remote $GIT_REPO_URL"
```

**Kubernetes:**
```bash
kubectl exec -n monitoring deployment/grafana-git-sync -- git ls-remote $GIT_REPO_URL
```

### Test Grafana connectivity

```bash
curl -u admin:admin http://localhost:3000/api/health
```

### Common Issues

1. **SSH key format errors:**
   - Ensure key is in OpenSSH format (not PEM)
   - Use `|` for multiline strings in YAML
   - Include header/footer lines

2. **Permission denied (publickey):**
   - Add public key to Git provider (GitHub/GitLab/Bitbucket)
   - Verify key format with `ssh-keygen -l -f key.pub`

3. **Grafana connection refused:**
   - Check `GRAFANA_URL` is correct
   - Verify Grafana is running: `curl http://grafana:3000/api/health`
   - For Docker: use service name (`http://grafana:3000`)
   - For K8s: use service name (`http://grafana.monitoring.svc.cluster.local:3000`)

4. **Dashboards not syncing:**
   - Check logs for errors
   - Verify JSON files are valid: `jq . dashboard.json`
   - Ensure Git branch is correct
   - Check `GIT_REPO_SUBDIR` if using subdirectory

---

## 📖 More Information

- [Configuration Guide](../docs/configuration.md)
- [Architecture](../docs/architecture.md)
- [Docker Deployment](../docs/deployment/docker.md)
- [Kubernetes Deployment](../docs/deployment/kubernetes.md)
- [Troubleshooting](../docs/troubleshooting.md)

---

**Need help?** Open an issue on [GitHub](https://github.com/efremov-it/grafana-git-sync/issues)
