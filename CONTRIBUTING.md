# Contributing to Grafana Git Sync

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

---

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Commit Message Guidelines](#commit-message-guidelines)

---

## 🤝 Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for everyone. Be respectful, collaborative, and professional.

### Our Standards

**Positive behavior:**
- Using welcoming and inclusive language
- Respecting differing viewpoints
- Accepting constructive criticism
- Focusing on what's best for the community

**Unacceptable behavior:**
- Harassment or discriminatory language
- Trolling or insulting comments
- Publishing private information without permission
- Other unprofessional conduct

---

## 🎯 How Can I Contribute?

### Reporting Bugs

Before creating a bug report:
1. Check [existing issues](https://github.com/efremov-it/grafana-git-sync/issues)
2. Verify you're using the latest version
3. Collect relevant information (logs, config, environment)

**Bug Report Template:**
```markdown
**Describe the bug**
A clear description of the bug.

**To Reproduce**
Steps to reproduce the behavior:
1. Set environment variables '...'
2. Run command '...'
3. See error

**Expected behavior**
What you expected to happen.

**Logs**
```
Paste relevant logs here
```

**Environment:**
- OS: [e.g., Linux, macOS]
- Deployment: [Docker, Kubernetes, Binary]
- Version: [e.g., v0.1.0]
- Go version: [e.g., 1.24]

**Additional context**
Any other relevant information.
```

### Suggesting Features

**Feature Request Template:**
```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other solutions you've thought about.

**Additional context**
Any other relevant information, mockups, examples.
```

### Contributing Code

1. **Check existing issues** - Avoid duplicate work
2. **Discuss major changes** - Open an issue first for big features
3. **Fork and clone** - Work in your own fork
4. **Create a branch** - Use descriptive branch names
5. **Make changes** - Follow coding guidelines
6. **Test thoroughly** - Write and run tests
7. **Submit PR** - Fill out the PR template

---

## 🛠 Development Setup

### Prerequisites

- Go 1.24 or later
- Git
- Docker (for testing)
- Kubernetes cluster (optional, for K8s testing)

### Clone Repository

```bash
git clone https://github.com/efremov-it/grafana-git-sync.git
cd grafana-git-sync
```

### Install Dependencies

```bash
go mod download
```

### Build

```bash
# Build binary
make build

# Or manually
go build -o bin/grafana-git-sync cmd/grafana-git-sync/main.go
```

### Run Locally

```bash
# Set environment variables
export GIT_REPO_URL=https://github.com/your-org/dashboards.git
export GIT_BRANCH=main
export GRAFANA_URL=http://localhost:3000
export GF_SECURITY_ADMIN_USER=admin
export GF_SECURITY_ADMIN_PASSWORD=admin

# Run
./bin/grafana-git-sync
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test -v ./pkg/git
```

### Build Docker Image

```bash
make docker-build

# Or manually
docker build -t grafana-git-sync:dev .
```

---

## 📝 Coding Guidelines

### Go Code Style

Follow standard Go conventions:

1. **Use `gofmt`:**
   ```bash
   gofmt -s -w .
   ```

2. **Use `golint`:**
   ```bash
   golint ./...
   ```

3. **Use `go vet`:**
   ```bash
   go vet ./...
   ```

4. **Run all checks:**
   ```bash
   make lint
   ```

### Code Organization

```
grafana-git-sync/
├── cmd/               # Entry points
│   └── grafana-git-sync/
│       └── main.go    # Main application
├── pkg/               # Reusable packages
│   ├── config/        # Configuration management
│   ├── git/           # Git operations
│   ├── grafana/       # Grafana API client
│   ├── health/        # Health check server
│   └── sync/          # Sync orchestration
├── docs/              # Documentation
├── examples/          # Example configurations
└── tests/             # Integration tests (future)
```

### Package Guidelines

1. **Single Responsibility:** Each package should have one clear purpose
2. **Minimal Dependencies:** Avoid circular dependencies
3. **Clear Interfaces:** Define interfaces for testability
4. **Error Handling:** Return errors, don't panic
5. **Context:** Use `context.Context` for cancellation

### Naming Conventions

```go
// Good
func NewGrafanaClient(url string) (*Client, error)
func (c *Client) GetDashboard(uid string) (*Dashboard, error)
var ErrDashboardNotFound = errors.New("dashboard not found")

// Bad
func grafana_client(u string) *Client  // Not exported, snake_case
func (c *Client) get(id string) error  // Unclear name
var err1 = errors.New("error")         // Non-descriptive
```

### Error Handling

```go
// Good: Return errors
func syncDashboard(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read file %s: %w", path, err)
    }
    // ...
}

// Bad: Panic
func syncDashboard(path string) {
    data := must(os.ReadFile(path))  // Don't panic!
    // ...
}
```

### Logging

Use standard log package or structured logging:

```go
import "log"

// Good
log.Printf("Syncing dashboard: %s (folder: %s)", name, folder)
log.Printf("ERROR: Failed to upload dashboard: %v", err)

// Future: Structured logging
logger.Info("dashboard synced",
    "name", name,
    "folder", folder,
    "duration_ms", duration.Milliseconds())
```

---

## 🧪 Testing

### Unit Tests

Write tests for all packages:

```go
// pkg/config/config_test.go
func TestLoadConfig(t *testing.T) {
    tests := []struct {
        name    string
        env     map[string]string
        want    *Config
        wantErr bool
    }{
        {
            name: "valid config",
            env: map[string]string{
                "GIT_REPO_URL": "https://github.com/test/repo.git",
                "GIT_BRANCH":   "main",
                "GRAFANA_URL":  "http://localhost:3000",
            },
            want: &Config{
                GitRepoURL: "https://github.com/test/repo.git",
                GitBranch:  "main",
                GrafanaURL: "http://localhost:3000",
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set environment
            for k, v := range tt.env {
                t.Setenv(k, v)
            }
            
            // Test
            got, err := LoadConfig()
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Test with real Grafana and Git:

```go
// tests/integration_test.go
// +build integration

func TestFullSync(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test Grafana
    grafanaURL := setupTestGrafana(t)
    defer teardownTestGrafana(t)
    
    // Setup test Git repo
    repoURL := setupTestRepo(t)
    defer teardownTestRepo(t)
    
    // Run sync
    // ...
    
    // Verify dashboards
    // ...
}
```

Run integration tests:
```bash
go test -tags=integration ./tests/...
```

### Test Coverage

Aim for >80% coverage:

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🔄 Pull Request Process

### Before Submitting

1. **Create an issue** (for non-trivial changes)
2. **Fork the repository**
3. **Create a feature branch:**
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. **Make your changes**
5. **Run tests:**
   ```bash
   make test
   make lint
   ```
6. **Commit with descriptive messages**
7. **Push to your fork:**
   ```bash
   git push origin feature/amazing-feature
   ```
8. **Open a Pull Request**

### PR Template

```markdown
**Description**
Brief description of changes.

**Related Issue**
Fixes #123

**Type of Change**
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature causing existing functionality to change)
- [ ] Documentation update

**Checklist**
- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Commented code, particularly complex areas
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added/updated
- [ ] All tests pass locally
- [ ] Dependent changes merged and published

**Testing**
Describe how you tested your changes.

**Screenshots** (if applicable)
Add screenshots for UI changes.
```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainer(s)
3. **Address feedback** - Update PR as needed
4. **Approval** - Maintainer approves PR
5. **Merge** - Maintainer merges to main

### After Merge

- Your contribution will be in the next release
- You'll be added to CONTRIBUTORS.md
- Thank you! 🎉

---

## 📝 Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation only
- **style:** Code style (formatting, no logic change)
- **refactor:** Code restructuring (no feature change)
- **perf:** Performance improvement
- **test:** Adding/updating tests
- **chore:** Maintenance (dependencies, build)

### Examples

```
feat(grafana): add dashboard versioning support

Implement Git commit metadata extraction and pass to Grafana API
as version message. Format: "commit <hash>: <message> - <author>"

Closes #42
```

```
fix(git): handle SSH key with CRLF line endings

Convert CRLF to LF when parsing SSH keys from environment variables.

Fixes #38
```

```
docs: add Kubernetes deployment guide

Add comprehensive K8s deployment documentation including:
- Deployment manifests
- Secret management options
- Health check configuration
- Troubleshooting tips
```

### Scope

Optional, indicates affected component:
- `git` - Git operations
- `grafana` - Grafana API client
- `sync` - Sync service
- `config` - Configuration
- `health` - Health checks
- `docs` - Documentation
- `docker` - Docker related
- `k8s` - Kubernetes related

---

## 🏷️ Release Process

Maintainers will:

1. Update CHANGELOG.md
2. Tag version: `vX.Y.Z`
3. Build release artifacts
4. Publish Docker image
5. Create GitHub release
6. Announce in discussions

Versioning follows [Semantic Versioning](https://semver.org/):
- **MAJOR:** Breaking changes
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes (backward compatible)

---

## 📞 Questions?

- **General questions:** [GitHub Discussions](https://github.com/efremov-it/grafana-git-sync/discussions)
- **Bug reports:** [GitHub Issues](https://github.com/efremov-it/grafana-git-sync/issues)
- **Feature requests:** [GitHub Issues](https://github.com/efremov-it/grafana-git-sync/issues)

---

**Thank you for contributing! 🙏**
