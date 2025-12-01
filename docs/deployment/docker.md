# Docker Deployment

Deploy Grafana Git Sync using Docker.

## Quick Start

### Using Docker Run

**SSH Authentication:**
```bash
docker run -d \
  --name grafana-git-sync \
  --network host \
  -p 8080:8080 \
  -e GIT_REPO_URL=ssh://git@github.com/your-org/dashboards.git \
  -e GIT_BRANCH=main \
  -e GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)" \
  -e GRAFANA_URL=http://localhost:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana-git-sync:latest
```

**HTTPS Authentication:**
```bash
docker run -d \
  --name grafana-git-sync \
  --network host \
  -p 8080:8080 \
  -e GIT_REPO_URL=https://github.com/your-org/dashboards.git \
  -e GIT_BRANCH=main \
  -e GIT_HTTPS_USER=bot-user \
  -e GIT_HTTPS_PASS=ghp_yourtoken \
  -e GRAFANA_URL=http://localhost:3000 \
  -e GF_SECURITY_TOKEN=glsa_yourtoken \
  grafana-git-sync:latest
```

## Docker Compose

See [examples/docker-compose.yml](../examples/docker-compose.yml) for complete example.

**Basic docker-compose.yml:**
```yaml
version: '3.8'

services:
  grafana-git-sync:
    image: grafana-git-sync:latest
    container_name: grafana-git-sync
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      GIT_REPO_URL: ssh://git@github.com/your-org/dashboards.git
      GIT_BRANCH: main
      GIT_SSH_KEY: |
        -----BEGIN OPENSSH PRIVATE KEY-----
        your_key_here
        -----END OPENSSH PRIVATE KEY-----
      GRAFANA_URL: http://grafana:3000
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: admin
      POLL_INTERVAL_SEC: 60
    healthcheck:
      test: ["CMD-SHELL", "wget --spider -q http://localhost:8080/healthz || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

**Start:**
```bash
docker-compose up -d
```

**View logs:**
```bash
docker-compose logs -f grafana-git-sync
```

## Building Custom Image

```bash
# Clone repository
git clone https://github.com/efremov-it/grafana-git-sync.git
cd grafana-git-sync

# Build image
docker build -t grafana-git-sync:latest .

# Or use Make
make docker-build
```

## Network Configuration

### Option 1: Host Network (Recommended for Local)
```yaml
network_mode: host
```
- Grafana accessible at `http://localhost:3000`
- Simplest configuration
- Works on Linux only

### Option 2: Bridge Network (Production)
```yaml
services:
  grafana:
    image: grafana/grafana:latest
    networks:
      - monitoring
  
  grafana-git-sync:
    image: grafana-git-sync:latest
    networks:
      - monitoring
    environment:
      GRAFANA_URL: http://grafana:3000  # Use service name
```

### Option 3: External Network
```yaml
services:
  grafana-git-sync:
    image: grafana-git-sync:latest
    networks:
      - external_network
    environment:
      GRAFANA_URL: https://grafana.example.com  # External URL
```

## Volume Mounts (Optional)

By default, everything runs in memory/temp dirs. For persistence:

```yaml
volumes:
  - ./git-repo:/data/git-repo
  - ./dashboards:/data/dashboards
environment:
  GIT_LOCAL_REPO_DIR: /data/git-repo
  DASHBOARDS_DIR: /data/dashboards
```

**Note:** Volumes are NOT required - stateless design is intentional.

## Health Check

Docker health check configuration:

```yaml
healthcheck:
  test: ["CMD-SHELL", "wget --spider -q http://localhost:8080/healthz || exit 1"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

Check health status:
```bash
docker inspect grafana-git-sync | grep -A 10 Health
curl http://localhost:8080/healthz | jq
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs grafana-git-sync

# Common issues:
# - Missing required env vars (GIT_REPO_URL, GIT_BRANCH, GRAFANA_URL)
# - Invalid SSH key format
# - Grafana not accessible
```

### SSH Key Issues
```bash
# Ensure key has no passphrase
ssh-keygen -p -f ~/.ssh/id_rsa

# Test key format
docker run -it --rm \
  -e GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)" \
  grafana-git-sync:latest
```

### Grafana Connection Failed
```bash
# Test from container
docker exec grafana-git-sync wget -O- http://grafana:3000/api/health
```

## Logs

View structured logs:
```bash
# Real-time
docker logs -f grafana-git-sync

# With timestamps
docker logs -f -t grafana-git-sync

# Last 100 lines
docker logs --tail 100 grafana-git-sync
```

**Log Emoji Indicators:**
- 🚀 Startup
- ✅ Success
- ❌ Error
- ⚠️ Warning
- 📦 New commit
- 📊 Stats
- 🔍 No changes
- 🔑 Authentication

## Security

### Secrets Management

**Docker Secrets (Swarm):**
```yaml
secrets:
  ssh_key:
    external: true

services:
  grafana-git-sync:
    secrets:
      - ssh_key
    environment:
      GIT_SSH_KEY_FILE: /run/secrets/ssh_key
```

**Environment Files:**
```bash
# .env file
GIT_REPO_URL=ssh://git@github.com/org/dashboards.git
GIT_SSH_KEY=...
GF_SECURITY_TOKEN=...

# Use with compose
docker-compose --env-file .env up -d
```

## Resource Limits

```yaml
services:
  grafana-git-sync:
    image: grafana-git-sync:latest
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
```

## Updating

```bash
# Pull latest image
docker pull grafana-git-sync:latest

# Restart container
docker-compose up -d

# Or
docker restart grafana-git-sync
```

**Note:** Stateless design means no data migration needed!
