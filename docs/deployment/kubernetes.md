# Kubernetes Deployment Guide

This guide covers deploying Grafana Git Sync on Kubernetes with various configuration options.

---

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Deployment Options](#deployment-options)
- [Configuration](#configuration)
- [Secrets Management](#secrets-management)
- [Health Checks](#health-checks)
- [Resource Management](#resource-management)
- [Troubleshooting](#troubleshooting)

---

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Docker image built or available in registry

### Basic Deployment

```bash
# Apply the complete manifest
kubectl apply -f examples/kubernetes.yaml

# Check deployment status
kubectl get pods -n monitoring

# View logs
kubectl logs -n monitoring deployment/grafana-git-sync -f

# Test health endpoint
kubectl port-forward -n monitoring svc/grafana-git-sync 8080:8080
curl http://localhost:8080/healthz
```

---

## 🎯 Deployment Options

### Option 1: Standalone Deployment

Deploy Grafana Git Sync as a standalone service connecting to external Grafana:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana-git-sync
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana-git-sync
  template:
    metadata:
      labels:
        app: grafana-git-sync
    spec:
      containers:
      - name: grafana-git-sync
        image: grafana-git-sync:latest
        env:
        - name: GRAFANA_URL
          value: "http://grafana.monitoring.svc.cluster.local:3000"
        # ... other env vars ...
```

### Option 2: Sidecar Container

Deploy as a sidecar alongside Grafana in the same pod:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      # Main Grafana container
      - name: grafana
        image: grafana/grafana:latest
        ports:
        - containerPort: 3000
        volumeMounts:
        - name: grafana-storage
          mountPath: /var/lib/grafana
      
      # Sidecar for Git sync
      - name: git-sync
        image: grafana-git-sync:latest
        env:
        - name: GRAFANA_URL
          value: "http://localhost:3000"  # Same pod
        - name: GIT_REPO_URL
          valueFrom:
            secretKeyRef:
              name: git-config
              key: repo-url
        - name: GIT_SSH_KEY
          valueFrom:
            secretKeyRef:
              name: git-ssh
              key: private-key
        - name: GF_SECURITY_ADMIN_USER
          valueFrom:
            secretKeyRef:
              name: grafana-creds
              key: username
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-creds
              key: password
        ports:
        - containerPort: 8080
          name: health
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
      
      volumes:
      - name: grafana-storage
        emptyDir: {}
```

### Option 3: Using Helm (Future)

```bash
# Add Helm repository
helm repo add grafana-git-sync https://charts.grafana-git-sync.io
helm repo update

# Install with values
helm install grafana-git-sync grafana-git-sync/grafana-git-sync \
  --namespace monitoring \
  --create-namespace \
  --set gitRepoUrl="ssh://git@github.com/your-org/dashboards.git" \
  --set gitBranch="main" \
  --set grafanaUrl="http://grafana:3000"
```

---

## ⚙️ Configuration

### Using ConfigMap

Store non-sensitive configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-git-sync-config
  namespace: monitoring
data:
  GIT_REPO_URL: "ssh://git@github.com/your-org/dashboards.git"
  GIT_BRANCH: "main"
  GIT_REPO_SUBDIR: "dashboards/production"  # Optional
  GRAFANA_URL: "http://grafana.monitoring.svc.cluster.local:3000"
  POLL_INTERVAL_SEC: "60"
  HEALTH_CHECK_PORT: "8080"
```

Reference in Deployment:

```yaml
envFrom:
- configMapRef:
    name: grafana-git-sync-config
```

### Using Secrets

Store sensitive data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: grafana-git-sync-secrets
  namespace: monitoring
type: Opaque
stringData:
  GIT_SSH_KEY: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
  GF_SECURITY_ADMIN_USER: admin
  GF_SECURITY_ADMIN_PASSWORD: supersecret
```

Reference in Deployment:

```yaml
envFrom:
- secretRef:
    name: grafana-git-sync-secrets
```

---

## 🔐 Secrets Management

### Option 1: Kubernetes Secrets (Basic)

```bash
# Create secret from file
kubectl create secret generic git-ssh-key \
  --from-file=ssh-key=~/.ssh/id_rsa \
  --namespace monitoring

# Create secret from literal
kubectl create secret generic grafana-creds \
  --from-literal=username=admin \
  --from-literal=password=admin123 \
  --namespace monitoring
```

### Option 2: Sealed Secrets

For GitOps workflows with public repositories:

```bash
# Install Sealed Secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Create sealed secret
echo -n "admin" | kubectl create secret generic grafana-creds \
  --dry-run=client \
  --from-file=username=/dev/stdin \
  --from-file=password=<(echo -n "admin123") \
  -o yaml | \
kubeseal -o yaml > sealed-secret.yaml

# Apply to cluster
kubectl apply -f sealed-secret.yaml
```

### Option 3: External Secrets Operator

Use with AWS Secrets Manager, HashiCorp Vault, etc.:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: grafana-git-sync-secrets
  namespace: monitoring
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: grafana-git-sync-secrets
    creationPolicy: Owner
  data:
  - secretKey: GIT_SSH_KEY
    remoteRef:
      key: grafana-git-sync/ssh-key
  - secretKey: GF_SECURITY_ADMIN_PASSWORD
    remoteRef:
      key: grafana-git-sync/admin-password
```

### Option 4: Vault Agent Injector

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana-git-sync
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "grafana-git-sync"
        vault.hashicorp.com/agent-inject-secret-ssh-key: "secret/data/grafana-git-sync/ssh-key"
        vault.hashicorp.com/agent-inject-template-ssh-key: |
          {{- with secret "secret/data/grafana-git-sync/ssh-key" -}}
          {{ .Data.data.key }}
          {{- end }}
    spec:
      serviceAccountName: grafana-git-sync
      containers:
      - name: grafana-git-sync
        image: grafana-git-sync:latest
        env:
        - name: GIT_SSH_KEY
          value: "/vault/secrets/ssh-key"
```

---

## 🏥 Health Checks

### Liveness Probe

Restarts container if unhealthy:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
  timeoutSeconds: 10
  failureThreshold: 3
```

### Readiness Probe

Removes pod from service if not ready:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Startup Probe

For slow-starting applications:

```yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 30  # 30 * 5 = 150 seconds max startup time
```

---

## 📊 Resource Management

### Resource Requests and Limits

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

### Horizontal Pod Autoscaler (HPA)

Not recommended (sync is stateful), but for high-availability:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: grafana-git-sync
  namespace: monitoring
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: grafana-git-sync
  minReplicas: 1
  maxReplicas: 3
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
```

**Note:** Multiple replicas may cause race conditions. Use leader election if needed.

### Pod Disruption Budget

Ensure availability during cluster maintenance:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: grafana-git-sync
  namespace: monitoring
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: grafana-git-sync
```

---

## 🔍 Monitoring and Observability

### Service Monitor (Prometheus)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: grafana-git-sync
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: grafana-git-sync
  endpoints:
  - port: health
    path: /metrics  # Future enhancement
    interval: 30s
```

### Logs

```bash
# Follow logs
kubectl logs -n monitoring deployment/grafana-git-sync -f

# Logs from specific pod
kubectl logs -n monitoring pod/grafana-git-sync-xxx -f

# Previous container logs (after restart)
kubectl logs -n monitoring deployment/grafana-git-sync --previous

# Logs with timestamps
kubectl logs -n monitoring deployment/grafana-git-sync --timestamps
```

### Events

```bash
# Watch events
kubectl get events -n monitoring --watch

# Filter by deployment
kubectl get events -n monitoring --field-selector involvedObject.name=grafana-git-sync
```

---

## 🔧 Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n monitoring -l app=grafana-git-sync
kubectl describe pod -n monitoring -l app=grafana-git-sync
```

### Exec into Container

```bash
kubectl exec -it -n monitoring deployment/grafana-git-sync -- sh

# Test Git connectivity
git ls-remote $GIT_REPO_URL

# Test Grafana connectivity
wget -O- $GRAFANA_URL/api/health

# Check environment variables
env | grep -E 'GIT|GRAFANA'
```

### Network Policies

If using network policies, allow traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: grafana-git-sync
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: grafana-git-sync
  policyTypes:
  - Egress
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow Grafana access
  - to:
    - podSelector:
        matchLabels:
          app: grafana
    ports:
    - protocol: TCP
      port: 3000
  # Allow Git (SSH)
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 22
  # Allow Git (HTTPS)
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

### Common Issues

**1. CrashLoopBackOff**

```bash
kubectl logs -n monitoring deployment/grafana-git-sync --previous
```

Common causes:
- Invalid SSH key format
- Wrong Grafana credentials
- Network connectivity issues

**2. ImagePullBackOff**

```bash
kubectl describe pod -n monitoring -l app=grafana-git-sync
```

Solutions:
- Check image name/tag
- Verify registry credentials
- Use imagePullSecrets if private registry

**3. Permission Denied (SSH)**

```bash
kubectl exec -n monitoring deployment/grafana-git-sync -- sh -c "ssh-keygen -l -f /dev/stdin <<< \"\$GIT_SSH_KEY\""
```

Solutions:
- Verify SSH key is in OpenSSH format
- Check public key is added to Git provider
- Test with `git ls-remote`

**4. Grafana Connection Refused**

```bash
kubectl exec -n monitoring deployment/grafana-git-sync -- wget -O- http://grafana:3000/api/health
```

Solutions:
- Check Grafana service name and namespace
- Verify Grafana is running
- Check network policies

---

## 📖 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Configuration Guide](../configuration.md)
- [Architecture](../architecture.md)
- [Docker Deployment](docker.md)
- [Troubleshooting](../troubleshooting.md)

---

**Need help?** Open an issue on [GitHub](https://github.com/efremov-it/grafana-git-sync/issues)
