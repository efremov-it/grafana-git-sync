# Architecture

## Overview

Grafana Git Sync is a stateless synchronization tool that automatically mirrors Grafana dashboards from a Git repository.

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│  Git Repo   │◄────────┤ Grafana-Git-Sync ├────────►│   Grafana   │
│ (Source of  │  Poll   │                  │  Upload │  (Target)   │
│   Truth)    │         │  ┌────────────┐  │         │             │
└─────────────┘         │  │ Health API │  │         └─────────────┘
                        │  └────────────┘  │
                        └──────────────────┘
                               :8080
```

## Core Components

### 1. Git Client (`pkg/git`)
**Responsibility:** Git operations

- **Clone** - Initial repository cloning
- **Pull** - Fetch latest changes
- **Commit Tracking** - Detect new commits
- **Metadata Extraction** - Get commit info (author, message, hash)

**Key Features:**
- SSH and HTTPS authentication
- Shallow clones (depth=1) for efficiency
- Commit metadata for versioning

### 2. Grafana Client (`pkg/grafana`)
**Responsibility:** Grafana API interactions

- **Token Management** - Auto-create service account tokens
- **Folder Operations** - Create nested folder structures
- **Dashboard Upload** - Upload with version metadata
- **Folder Caching** - Avoid duplicate folder creation

**Key Features:**
- Unlimited folder nesting via `parentUid` parameter
- Folder reuse detection
- Version message support

### 3. Sync Service (`pkg/sync`)
**Responsibility:** Dashboard file handling

- **File Discovery** - Find all `.json` files
- **Folder Graph Building** - Map directory structure
- **Dashboard Loading** - Parse JSON files
- **Change Detection** - Hash-based change tracking

**Key Features:**
- Preserves directory hierarchy
- Smart sync (hash tracking)
- Subdirectory support

### 4. Health Checker (`pkg/health`)
**Responsibility:** Health monitoring

- **HTTP Server** - Provides `/healthz` endpoint
- **Status Tracking** - Monitor Grafana and Git connectivity
- **Metrics** - Last sync time, error messages

**Status Values:**
- `healthy` - All systems operational
- `degraded` - One system down
- `unhealthy` - Both systems down

### 5. Config (`pkg/config`)
**Responsibility:** Configuration management

- Load environment variables
- Validate required settings
- Provide sensible defaults

## Data Flow

### Initial Sync
```
1. Load Config
2. Start Health Server (:8080)
3. Connect to Grafana → Create/Validate Token
4. Clone Git Repository
5. Build Folder Structure → Create Folders in Grafana
6. Load All Dashboards → Upload to Grafana with Version Info
7. Mark Initial Sync Complete
```

### Continuous Sync Loop
```
while true:
  1. Fetch Latest Commit
  2. Compare with Last Known Commit
  
  if NEW COMMIT:
    3. Get Commit Metadata (author, message)
    4. Copy Dashboards from Repo
    5. Detect Changed Files (hash comparison)
    6. Build Folder Graph
    7. Create Missing Folders
    8. Upload Changed Dashboards (with version message)
    9. Update Health Status
    10. Save Last Commit Hash
  
  Sleep POLL_INTERVAL_SEC
```

## Folder Management

### Folder Graph Structure
```go
type FolderNode struct {
    Name     string
    FullPath string
    ID       int         // Grafana folder ID
    UID      string      // Grafana folder UID
    Children []*FolderNode
}
```

### Folder Creation Algorithm
```
For each directory in Git:
  1. Check if folder exists in Grafana (by title + parentUid)
  2. If exists: Reuse existing folder
  3. If not: Create new folder with parentUid
  4. Cache folder ID and UID
  5. Recursively process children
```

**Example:**
```
dashboards/
  ├── db/          → Create "db" (parent='')
  │   └── mysql/   → Create "mysql" (parent=UID of 'db')
  └── infra/       → Create "infra" (parent='')
```

## Dashboard Versioning

### Version Message Format
```
commit {short_hash}: {commit_message} - {author}
```

### API Payload
```json
{
  "dashboard": { /* dashboard JSON */ },
  "folderId": 123,
  "overwrite": true,
  "message": "commit abc1234: Updated metrics - John Doe"
}
```

### Grafana Storage
- Version stored in Grafana database
- Accessible via Grafana UI → Dashboard Settings → Versions
- Can diff and restore previous versions

## Stateless Design

**No Persistent State:**
- No database
- No local file storage (beyond temp dirs)
- File hashes stored in memory only

**Implications:**
- Restart triggers full dashboard re-upload (by design)
- Horizontal scaling friendly
- No state corruption risk
- Container-native

**Why Stateless?**
- ✅ Simple deployment
- ✅ No backup/restore needed
- ✅ Kubernetes-friendly
- ✅ No state synchronization issues
- ✅ Re-upload is cheap (idempotent)

## Performance Considerations

### Optimization Strategies
1. **Shallow Clone** - `depth=1` reduces clone time
2. **Smart Sync** - Only upload changed dashboards (in-memory hash tracking)
3. **Folder Caching** - Avoid redundant API calls
4. **Single Branch** - Only sync specified branch

### Scalability
- **Small repos** (<100 dashboards): Sub-second sync
- **Medium repos** (100-1000 dashboards): 1-5 seconds
- **Large repos** (1000+ dashboards): 5-30 seconds

Initial sync is slowest (uploads all). Subsequent syncs are faster (only changes).

## Error Handling

### Failure Modes
1. **Git Fetch Failure** - Retries on next poll interval
2. **Grafana API Failure** - Logs error, continues with other dashboards
3. **Dashboard Parse Error** - Skips dashboard, continues with others

### Health Status
- Errors update health endpoint status
- Last error message tracked
- Degraded mode allows continued operation

## Security

### Secrets Management
- Environment variables for all credentials
- No secrets in logs (passwords redacted)
- SSH keys in memory only
- Service account tokens (not user passwords)

### Grafana Permissions
- Requires Admin role for service account creation
- OR provide pre-created service account token
- Uses token-based auth (not session cookies)

## Extensibility

### Future Enhancements
- **Webhook Mode** - React to Git webhooks instead of polling
- **Prometheus Metrics** - Export sync metrics
- **Dashboard Deletion** - Remove dashboards deleted from Git
- **Diff View** - Show what changed before syncing
