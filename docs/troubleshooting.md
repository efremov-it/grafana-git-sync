# Troubleshooting Guide

Common issues and solutions for Grafana Git Sync.

---

## 📋 Table of Contents

- [General Debugging](#general-debugging)
- [Git Issues](#git-issues)
- [Grafana Issues](#grafana-issues)
- [Dashboard Issues](#dashboard-issues)
- [Docker Issues](#docker-issues)
- [Kubernetes Issues](#kubernetes-issues)
- [Performance Issues](#performance-issues)

---

## 🔍 General Debugging

### Enable Verbose Logging

Check application logs for detailed information:

**Docker:**
```bash
docker logs grafana-git-sync -f
```

**Kubernetes:**
```bash
kubectl logs -n monitoring deployment/grafana-git-sync -f
```

### Check Health Endpoint

```bash
curl http://localhost:8080/healthz | jq
```

Expected response:
```json
{
  "status": "healthy",
  "grafana_connected": true,
  "git_connected": true,
  "last_sync": "2024-01-15T10:30:00Z",
  "sync_interval_sec": 60
}
```

### Verify Environment Variables

**Docker:**
```bash
docker exec grafana-git-sync env | grep -E 'GIT|GRAFANA'
```

**Kubernetes:**
```bash
kubectl exec -n monitoring deployment/grafana-git-sync -- env | grep -E 'GIT|GRAFANA'
```

---

## 🔐 Git Issues

### Problem: Permission denied (publickey)

**Symptoms:**
```
ERROR: Permission denied (publickey)
fatal: Could not read from remote repository
```

**Causes:**
1. SSH key not added to Git provider
2. Wrong SSH key format
3. SSH key permissions too open
4. Wrong repository URL

**Solutions:**

1. **Verify SSH key format (OpenSSH, not PEM):**
   ```bash
   head -1 ~/.ssh/id_rsa
   # Should show: -----BEGIN OPENSSH PRIVATE KEY-----
   # NOT: -----BEGIN RSA PRIVATE KEY-----
   ```

2. **Convert PEM to OpenSSH:**
   ```bash
   ssh-keygen -p -m OpenSSH -f ~/.ssh/id_rsa
   ```

3. **Test SSH key:**
   ```bash
   # GitHub
   ssh -T git@github.com
   
   # GitLab
   ssh -T git@gitlab.com
   
   # Bitbucket
   ssh -T git@bitbucket.org
   ```

4. **Add public key to Git provider:**
   - GitHub: Settings → SSH and GPG keys → New SSH key
   - GitLab: Preferences → SSH Keys
   - Bitbucket: Personal settings → SSH keys

5. **Check key in container:**
   ```bash
   docker exec grafana-git-sync sh -c 'echo "$GIT_SSH_KEY" | ssh-keygen -l -f /dev/stdin'
   ```

### Problem: Repository not found

**Symptoms:**
```
ERROR: Repository not found
fatal: repository 'https://github.com/org/repo.git/' not found
```

**Solutions:**

1. **Verify repository URL:**
   ```bash
   # SSH format
   GIT_REPO_URL=ssh://git@github.com/org/repo.git
   
   # HTTPS format
   GIT_REPO_URL=https://github.com/org/repo.git
   ```

2. **Check repository access:**
   ```bash
   git ls-remote $GIT_REPO_URL
   ```

3. **For private repos, ensure authentication is configured:**
   - SSH: `GIT_SSH_KEY` is set
   - HTTPS: `GIT_HTTPS_USER` and `GIT_HTTPS_PASS` are set

### Problem: Git clone fails with SSL error

**Symptoms:**
```
ERROR: SSL certificate problem: unable to get local issuer certificate
```

**Solutions:**

1. **Use SSH instead of HTTPS**
2. **For self-hosted Git, add CA certificate** (future enhancement)

### Problem: Subdirectory not found

**Symptoms:**
```
ERROR: Subdirectory 'dashboards' not found in repository
```

**Solutions:**

1. **Verify subdirectory path:**
   ```bash
   # List repository structure
   git clone $GIT_REPO_URL /tmp/test-repo
   ls -la /tmp/test-repo
   ```

2. **Check `GIT_REPO_SUBDIR` value:**
   ```bash
   # Should be relative to repo root
   GIT_REPO_SUBDIR=dashboards/production
   # NOT: /dashboards/production
   ```

---

## 🎯 Grafana Issues

### Problem: Unauthorized (401)

**Symptoms:**
```
ERROR: Grafana API error: 401 Unauthorized
```

**Causes:**
1. Wrong admin credentials
2. Invalid service token
3. Token expired

**Solutions:**

1. **Test Grafana credentials:**
   ```bash
   curl -u admin:admin http://localhost:3000/api/health
   ```

2. **Verify environment variables:**
   ```bash
   # Option 1: Admin credentials
   GF_SECURITY_ADMIN_USER=admin
   GF_SECURITY_ADMIN_PASSWORD=admin
   
   # Option 2: Service token
   GF_SECURITY_TOKEN=glsa_...
   ```

3. **Create new service account token:**
   - Login to Grafana
   - Administration → Service Accounts
   - Create service account with Admin role
   - Generate token and use in `GF_SECURITY_TOKEN`

### Problem: Forbidden (403)

**Symptoms:**
```
ERROR: Grafana API error: 403 Forbidden
```

**Causes:**
1. Service account has insufficient permissions
2. Organization mismatch

**Solutions:**

1. **Check service account role:**
   - Must be Admin or Editor with dashboard permissions
   - Cannot be Viewer

2. **Verify organization:**
   ```bash
   curl -u admin:admin http://localhost:3000/api/org
   ```

### Problem: Connection refused

**Symptoms:**
```
ERROR: dial tcp 127.0.0.1:3000: connect: connection refused
```

**Causes:**
1. Wrong Grafana URL
2. Grafana not running
3. Network issues

**Solutions:**

1. **Check Grafana is running:**
   ```bash
   curl http://localhost:3000/api/health
   ```

2. **Verify URL in Docker network:**
   ```bash
   # Docker Compose: use service name
   GRAFANA_URL=http://grafana:3000
   
   # Kubernetes: use service FQDN
   GRAFANA_URL=http://grafana.monitoring.svc.cluster.local:3000
   
   # Host network: use localhost
   GRAFANA_URL=http://localhost:3000
   ```

3. **Test from container:**
   ```bash
   docker exec grafana-git-sync wget -O- $GRAFANA_URL/api/health
   ```

---

## 📊 Dashboard Issues

### Problem: Dashboards not appearing in Grafana

**Symptoms:**
- Logs show "Dashboard uploaded successfully"
- Dashboards don't appear in Grafana UI

**Solutions:**

1. **Check Grafana search:**
   - Search for dashboard by name
   - Check specific folder

2. **Verify dashboard JSON:**
   ```bash
   jq . dashboard.json
   ```

3. **Check dashboard UID conflicts:**
   ```bash
   # Each dashboard must have unique UID
   grep -r '"uid"' dashboards/
   ```

### Problem: Duplicate folders created

**Symptoms:**
- Same folder appears multiple times
- Dashboards in wrong folders

**Solutions:**

This should be fixed in v0.1.0+. If you still see duplicates:

1. **Delete duplicate folders manually in Grafana**
2. **Restart sync service:**
   ```bash
   docker restart grafana-git-sync
   ```

3. **Check logs for folder creation:**
   ```bash
   docker logs grafana-git-sync | grep -i folder
   ```

### Problem: Dashboard versions not showing Git commits

**Symptoms:**
- Dashboard versions exist but don't show Git commit info

**Solutions:**

1. **Verify Dashboard Versioning is enabled (v0.1.0+):**
   ```bash
   docker logs grafana-git-sync | grep "commit"
   ```

2. **Check Git commit history:**
   ```bash
   git log --oneline dashboards/
   ```

3. **Ensure dashboard JSON changes are committed:**
   ```bash
   git diff dashboards/
   git add dashboards/
   git commit -m "Update dashboard"
   ```

### Problem: Invalid dashboard JSON

**Symptoms:**
```
ERROR: Invalid dashboard JSON: unexpected end of JSON input
```

**Solutions:**

1. **Validate JSON syntax:**
   ```bash
   find dashboards/ -name "*.json" -exec jq . {} \;
   ```

2. **Check for common issues:**
   - Missing closing braces
   - Trailing commas
   - Invalid escape sequences

3. **Export from Grafana and compare:**
   - Export working dashboard
   - Compare structure

---

## 🐳 Docker Issues

### Problem: Container exits immediately

**Symptoms:**
```
docker ps
# Container not listed
docker ps -a
# Shows Exited (1)
```

**Solutions:**

1. **Check logs:**
   ```bash
   docker logs grafana-git-sync
   ```

2. **Common causes:**
   - Missing required environment variables
   - Invalid SSH key format
   - Wrong repository URL

3. **Test with minimal config:**
   ```bash
   docker run --rm \
     -e GIT_REPO_URL=https://github.com/public/repo.git \
     -e GIT_BRANCH=main \
     -e GRAFANA_URL=http://localhost:3000 \
     -e GF_SECURITY_ADMIN_USER=admin \
     -e GF_SECURITY_ADMIN_PASSWORD=admin \
     grafana-git-sync:latest
   ```

### Problem: Network issues in Docker

**Symptoms:**
```
ERROR: dial tcp: lookup grafana: no such host
```

**Solutions:**

1. **Use correct network:**
   ```bash
   # Option 1: Host network
   docker run --network=host ...
   
   # Option 2: Same network as Grafana
   docker network create monitoring
   docker run --network=monitoring ...
   
   # Option 3: Docker Compose (automatic)
   docker-compose up
   ```

2. **Verify service names:**
   ```bash
   docker network inspect monitoring
   ```

### Problem: SSH key with newlines in Docker

**Symptoms:**
```
ERROR: invalid key format
```

**Solutions:**

Use literal newlines in docker-compose.yml:

```yaml
environment:
  GIT_SSH_KEY: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAA...
    -----END OPENSSH PRIVATE KEY-----
```

For `docker run`, use `$(cat key)`:

```bash
docker run -e GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)" ...
```

---

## ☸️ Kubernetes Issues

### Problem: CrashLoopBackOff

**Symptoms:**
```
kubectl get pods
# STATUS: CrashLoopBackOff
```

**Solutions:**

1. **Check logs from previous run:**
   ```bash
   kubectl logs -n monitoring pod/grafana-git-sync-xxx --previous
   ```

2. **Describe pod for events:**
   ```bash
   kubectl describe pod -n monitoring grafana-git-sync-xxx
   ```

3. **Common causes:**
   - Invalid secret data
   - Wrong environment variables
   - Resource constraints

### Problem: ImagePullBackOff

**Symptoms:**
```
kubectl get pods
# STATUS: ImagePullBackOff
```

**Solutions:**

1. **Check image name:**
   ```yaml
   image: grafana-git-sync:latest  # Verify tag exists
   ```

2. **For private registry, add imagePullSecrets:**
   ```yaml
   spec:
     imagePullSecrets:
     - name: regcred
   ```

3. **Create registry secret:**
   ```bash
   kubectl create secret docker-registry regcred \
     --docker-server=registry.example.com \
     --docker-username=user \
     --docker-password=pass \
     --namespace=monitoring
   ```

### Problem: Secret not found

**Symptoms:**
```
ERROR: Error from server (NotFound): secrets "grafana-git-sync-ssh" not found
```

**Solutions:**

1. **List secrets:**
   ```bash
   kubectl get secrets -n monitoring
   ```

2. **Create missing secret:**
   ```bash
   kubectl create secret generic grafana-git-sync-ssh \
     --from-file=ssh-key=~/.ssh/id_rsa \
     --namespace=monitoring
   ```

3. **Verify secret data:**
   ```bash
   kubectl get secret -n monitoring grafana-git-sync-ssh -o yaml
   ```

---

## ⚡ Performance Issues

### Problem: High CPU usage

**Symptoms:**
- Container using excessive CPU
- Slow dashboard sync

**Solutions:**

1. **Increase poll interval:**
   ```bash
   POLL_INTERVAL_SEC=300  # 5 minutes instead of 60 seconds
   ```

2. **Check repository size:**
   ```bash
   du -sh /tmp/git-repo
   ```

3. **Use subdirectory for large monorepos:**
   ```bash
   GIT_REPO_SUBDIR=dashboards  # Only sync dashboards/
   ```

4. **Set resource limits (Kubernetes):**
   ```yaml
   resources:
     limits:
       cpu: 500m
       memory: 256Mi
   ```

### Problem: Slow dashboard sync

**Symptoms:**
- Takes several minutes to sync dashboards
- Timeout errors

**Solutions:**

1. **Check number of dashboards:**
   ```bash
   find dashboards/ -name "*.json" | wc -l
   ```

2. **Smart sync reduces uploads (v0.1.0+):**
   - Only changed dashboards are uploaded
   - Check logs for "No changes detected"

3. **Optimize dashboard JSON:**
   - Remove unnecessary data
   - Simplify complex queries

---

## 📖 Additional Resources

- [Configuration Guide](configuration.md)
- [Architecture](architecture.md)
- [Docker Deployment](deployment/docker.md)
- [Kubernetes Deployment](deployment/kubernetes.md)
- [GitHub Issues](https://github.com/efremov-it/grafana-git-sync/issues)

---

**Still having issues?** Open a [GitHub Issue](https://github.com/efremov-it/grafana-git-sync/issues) with:
- Logs from the application
- Your configuration (redact secrets!)
- Steps to reproduce
- Expected vs actual behavior
