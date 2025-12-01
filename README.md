# Grafana Git Sync

**Git → Sync → Grafana API**

Automatically synchronize Grafana dashboards from a Git repository with real-time updates on every commit.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

## ✨ Key Features

- **📁 Unlimited Folder Nesting** - Preserves full directory hierarchy from Git
- **🔄 Continuous Sync** - Auto-detects Git commits and syncs instantly
- **📝 Dashboard Versioning** - Links Grafana versions to Git commits (author, message)
- **🚀 Smart Sync** - Only uploads changed dashboards
- **🏥 Health Checks** - HTTP endpoint for Docker/Kubernetes probes
- **🔐 Flexible Auth** - SSH or HTTPS for Git, tokens or admin creds for Grafana
- **🐳 Stateless** - No persistent state, container-native design

---

## 🚀 Quick Start

### Docker Run

```bash
docker run -d \
  --name grafana-git-sync \
  -p 8080:8080 \
  -e GIT_REPO_URL=ssh://git@github.com/your-org/dashboards.git \
  -e GIT_BRANCH=main \
  -e GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)" \
  -e GRAFANA_URL=http://localhost:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana-git-sync:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  grafana-git-sync:
    image: grafana-git-sync:latest
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
```

**More examples:** [examples/](examples/)

---

## 📋 Configuration

### Required Environment Variables

| Variable | Description |
|----------|-------------|
| `GIT_REPO_URL` | Git repository URL (SSH or HTTPS) |
| `GIT_BRANCH` | Branch or tag to sync |
| `GRAFANA_URL` | Grafana instance URL |

### Authentication (Git)

**SSH:**
```bash
GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)"
```

**HTTPS:**
```bash
GIT_HTTPS_USER=username
GIT_HTTPS_PASS=token_or_password
```

### Authentication (Grafana)

**Auto-create token:**
```bash
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=admin
```

**Use existing token:**
```bash
GF_SECURITY_TOKEN=glsa_your_token
```

### Optional Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `POLL_INTERVAL_SEC` | `60` | Git polling interval |
| `GIT_REPO_SUBDIR` | `.` | Subdirectory with dashboards |
| `HEALTH_CHECK_PORT` | `8080` | Health endpoint port |

**Full configuration reference:** [docs/configuration.md](docs/configuration.md)

---

## 📚 Documentation

- **[Configuration Guide](docs/configuration.md)** - All environment variables explained
- **[Architecture](docs/architecture.md)** - How it works internally
- **[Docker Deployment](docs/deployment/docker.md)** - Docker and Docker Compose
- **[Kubernetes Deployment](docs/deployment/kubernetes.md)** - K8s manifests and Helm
- **[Systemd Deployment](docs/deployment/systemd.md)** - Linux systemd service
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

---

## 🔧 How It Works

1. **Clone Git Repo** - Downloads dashboard JSON files
2. **Build Folder Structure** - Maps Git directories to Grafana folders
3. **Upload Dashboards** - Syncs to Grafana with version metadata
4. **Poll for Changes** - Checks Git every N seconds
5. **Smart Sync** - Only uploads modified dashboards

**Architecture details:** [docs/architecture.md](docs/architecture.md)

---

## 🎯 Why Use This?

**No other OSS solution offers:**
- ✅ API-based sync (no Grafana provisioning or restarts)
- ✅ Real-time updates on Git commits
- ✅ Unlimited folder nesting
- ✅ Dashboard versioning linked to Git
- ✅ Kubernetes-native (stateless, health checks)
- ✅ Zero configuration files (pure env vars)

**Perfect for:**
- GitOps workflows
- Multi-environment dashboard management
- Team collaboration with Git-based review
- Disaster recovery (Git as source of truth)

---

## 🛠 Build from Source

```bash
git clone https://github.com/efremov-it/grafana-git-sync.git
cd grafana-git-sync

# Build binary
make build

# Build Docker image
make docker-build

# Run
./bin/grafana-git-sync
```

---

## 🗺 Roadmap

- [x] Dashboard versioning (v0.1.0)
- [x] Health check endpoint (v0.1.0)
- [x] Smart sync (v0.1.0)
- [x] Unlimited folder nesting (v0.1.0)
- [ ] Dashboard deletion sync
- [ ] Webhook mode
- [ ] Prometheus metrics
- [ ] Helm chart

---

## 📝 License

MIT License - see [LICENSE](LICENSE)

---

## 🤝 Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open Pull Request

---

## 📧 Support

- 🐛 [GitHub Issues](https://github.com/efremov-it/grafana-git-sync/issues)
- 💬 [GitHub Discussions](https://github.com/efremov-it/grafana-git-sync/discussions)
- 📖 [Documentation](docs/)

---

**Made with ❤️ for the Grafana community**
