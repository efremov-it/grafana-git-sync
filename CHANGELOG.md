# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-12-01

### Added
- **Dashboard Versioning** ✨ - Links Grafana dashboard versions to Git commits with author and message
- **Health Check Endpoint** - HTTP server on port 8080 with `/healthz` endpoint for Docker/Kubernetes probes
- **Smart Sync** - Tracks file hashes to only upload modified dashboards (reduces API load)
- **Unlimited Folder Nesting** - Supports any depth of nested folders (fixed parentUid API usage)
- **Folder Reuse** - Intelligent folder detection to avoid duplicate folder creation
- **Automatic Token Management** - Creates Grafana service account tokens automatically
- **SSH Authentication** - Support for Git SSH key authentication
- **HTTPS Authentication** - Support for Git HTTPS username/password authentication
- **Subdirectory Support** - Sync dashboards from any subdirectory via `GIT_REPO_SUBDIR`
- **Continuous Synchronization** - Polls Git repository and syncs on new commits
- **Stateless Design** - No persistent state, container-friendly

### Documentation
- Comprehensive README with quick start examples
- Russian translation (README.ru.md)
- Health check configuration examples
- Docker and Docker Compose examples
- Kubernetes deployment manifests and guide
- Systemd service unit file and deployment guide
- GitHub Actions automated release workflow

### Technical
- Go 1.24
- Built with go-git for Git operations
- RESTful Grafana API integration
- Structured logging with emoji indicators
- Full test coverage for core packages

## [Unreleased]

### Planned
- Dashboard deletion when removed from Git
- Webhook mode for instant updates
- Helm chart for Kubernetes
- Prometheus metrics endpoint
- Sidecar deployment documentation

[0.1.0]: https://github.com/efremov-it/grafana-git-sync/releases/tag/v0.1.0
