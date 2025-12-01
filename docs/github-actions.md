# GitHub Actions Release Guide

This guide explains the automated release process using GitHub Actions.

## 🚀 Automated Release Process

When you push a tag (e.g., `v0.1.0`), GitHub Actions automatically:

1. ✅ Builds binaries for multiple platforms (Linux, macOS, Windows)
2. ✅ Builds and pushes Docker images (multi-arch: amd64, arm64)
3. ✅ Creates GitHub Release with changelog
4. ✅ Uploads binaries to GitHub Releases
5. ✅ Updates documentation with version numbers

---

## 📋 Prerequisites

### 1. GitHub Repository Secrets

Configure these secrets in your repository settings (`Settings` → `Secrets and variables` → `Actions`):

#### Optional (for Docker Hub):
- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `DOCKERHUB_TOKEN` - Docker Hub access token

**Note:** GitHub Container Registry (ghcr.io) works automatically without additional secrets.

### 2. Repository Permissions

Ensure GitHub Actions has write permissions:
- Go to `Settings` → `Actions` → `General`
- Under "Workflow permissions", select "Read and write permissions"
- Check "Allow GitHub Actions to create and approve pull requests"

---

## 🏷️ Creating a Release

### Step 1: Update CHANGELOG.md

```bash
# Edit CHANGELOG.md and add new version section
vim CHANGELOG.md
```

Example:
```markdown
## [0.2.0] - 2025-01-15

### Added
- New feature X
- Enhancement Y

### Fixed
- Bug fix Z

### Changed
- Updated dependency A
```

### Step 2: Commit Changes

```bash
git add CHANGELOG.md
git commit -m "docs: prepare v0.2.0 release"
git push origin main
```

### Step 3: Create and Push Tag

```bash
# Create annotated tag
git tag -a v0.2.0 -m "Release version 0.2.0"

# Push tag to trigger release workflow
git push origin v0.2.0
```

### Step 4: Monitor Release

1. Go to GitHub → `Actions` tab
2. Watch the "Release" workflow execution
3. Check for any errors in the workflow logs
4. Once complete, go to `Releases` tab to see the new release

---

## 📦 What Gets Built

### Binaries

Built for these platforms:
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64
- **Windows**: amd64

Binary names:
- `grafana-git-sync-linux-amd64`
- `grafana-git-sync-linux-arm64`
- `grafana-git-sync-darwin-amd64`
- `grafana-git-sync-darwin-arm64`
- `grafana-git-sync-windows-amd64.exe`

### Docker Images

Published to:

**GitHub Container Registry (automatic):**
```
ghcr.io/efremov-it/grafana-git-sync:latest
ghcr.io/efremov-it/grafana-git-sync:0.2.0
ghcr.io/efremov-it/grafana-git-sync:0.2
ghcr.io/efremov-it/grafana-git-sync:0
```

**Docker Hub (if configured):**
```
yourusername/grafana-git-sync:latest
yourusername/grafana-git-sync:0.2.0
yourusername/grafana-git-sync:0.2
yourusername/grafana-git-sync:0
```

Architectures:
- linux/amd64
- linux/arm64

---

## 🔧 Workflow Configuration

### Release Workflow (.github/workflows/release.yml)

Triggers on tag push matching `v*` pattern:

```yaml
on:
  push:
    tags:
      - 'v*'
```

Jobs:
1. **build-binaries** - Compiles Go binaries for all platforms
2. **build-docker** - Builds and pushes Docker images
3. **create-release** - Creates GitHub Release with changelog
4. **update-docs** - Updates version numbers in documentation

### Build Workflow (.github/workflows/build.yml)

Runs on every push to main/develop and pull requests:

Jobs:
1. **test** - Runs unit tests with coverage
2. **lint** - Runs golangci-lint
3. **build** - Builds binaries for all platforms
4. **docker** - Tests Docker image build

---

## 🐳 Docker Hub Setup (Optional)

If you want to publish to Docker Hub:

### 1. Create Docker Hub Access Token

1. Login to Docker Hub
2. Go to Account Settings → Security
3. Click "New Access Token"
4. Name: `GitHub Actions`
5. Permissions: Read, Write, Delete
6. Copy the token

### 2. Add Secrets to GitHub

1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add `DOCKERHUB_USERNAME` with your Docker Hub username
4. Add `DOCKERHUB_TOKEN` with the access token

### 3. Update Workflow (Optional)

The workflow automatically detects if Docker Hub secrets exist. If they do, it pushes to both registries.

To push only to Docker Hub, edit `.github/workflows/release.yml`:

```yaml
- name: Extract metadata
  id: meta
  uses: docker/metadata-action@v5
  with:
    images: |
      yourusername/grafana-git-sync
    tags: |
      type=semver,pattern={{version}}
      type=semver,pattern={{major}}.{{minor}}
      type=semver,pattern={{major}}
      type=raw,value=latest
```

---

## 📝 Changelog Automation

The release workflow automatically:

1. Extracts changelog from `CHANGELOG.md` for the current version
2. Generates commit list since last tag
3. Includes both in the GitHub Release notes

### Changelog Format

Use [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [Unreleased]

### Added
- Feature in development

## [0.2.0] - 2025-01-15

### Added
- New feature X

### Fixed
- Bug fix Y

[Unreleased]: https://github.com/efremov-it/grafana-git-sync/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/efremov-it/grafana-git-sync/releases/tag/v0.2.0
```

---

## 🔄 Version Updates

The workflow automatically updates version references in:

1. `docs/deployment/systemd.md` - Download commands
2. `CHANGELOG.md` - Adds release date

These changes are committed back to the main branch automatically.

---

## 🐛 Troubleshooting

### Workflow Fails on "Permission Denied"

**Solution:** Enable write permissions for GitHub Actions:
```
Settings → Actions → General → Workflow permissions → Read and write permissions
```

### Docker Push Fails

**Solution:** Check secrets are configured correctly:
```bash
# Test Docker Hub login locally
echo $DOCKERHUB_TOKEN | docker login -u $DOCKERHUB_USERNAME --password-stdin
```

### Binary Build Fails

**Solution:** Check Go version compatibility:
- Ensure `go.mod` specifies correct Go version
- Update workflow if needed: `.github/workflows/release.yml`

### Changelog Not Extracted

**Solution:** Verify CHANGELOG.md format:
```markdown
## [0.2.0] - 2025-01-15  ← Must match this format exactly
```

---

## 🎯 Best Practices

### Semantic Versioning

Follow [SemVer](https://semver.org/):
- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v0.2.0): New features, backward compatible
- **PATCH** (v0.1.1): Bug fixes, backward compatible

### Pre-releases

For beta/RC versions:
```bash
git tag -a v0.2.0-beta.1 -m "Beta release 0.2.0-beta.1"
git push origin v0.2.0-beta.1
```

Workflow detects pre-release tags and marks them accordingly.

### Release Checklist

Before creating a tag:
- [ ] Update CHANGELOG.md with all changes
- [ ] Update version in code if hardcoded
- [ ] Run tests locally: `make test`
- [ ] Build locally: `make build`
- [ ] Test Docker build: `make docker-build`
- [ ] Review diff: `git diff main`
- [ ] Commit and push to main
- [ ] Create tag and push
- [ ] Monitor GitHub Actions
- [ ] Verify release artifacts
- [ ] Test Docker image pull
- [ ] Announce release

---

## 📖 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [GitHub Release Action](https://github.com/softprops/action-gh-release)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)

---

**Need help?** Open an issue on [GitHub](https://github.com/efremov-it/grafana-git-sync/issues)
