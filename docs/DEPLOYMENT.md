# Prysm Deployment Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Deployment Modes](#deployment-modes)
3. [Standalone Deployment](#standalone-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Distributed Deployment with NATS](#distributed-deployment-with-nats)
6. [Kubernetes Sidecar Injection](#kubernetes-sidecar-injection)
7. [Configuration Management](#configuration-management)
8. [Production Considerations](#production-considerations)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements
- **Operating System**: Linux (Ubuntu 20.04+, RHEL 8+, or similar)
- **Go Version**: 1.26+ (for building from source)
- **Architecture**: amd64, arm64

### For Disk Health Metrics
- `smartmontools` installed (`smartctl` command)
- `nvme-cli` for NVMe devices
- Root or appropriate permissions to access disk devices

### For Operations Log Producer
- Read access to Ceph RadosGW log files
- Write access for log rotation

### For Kubernetes Deployment
- Kubernetes 1.24+
- `cert-manager` (for webhook TLS certificates)
- Rook-Ceph operator (for RGW integration)

### For NATS Integration
- NATS Server 2.12+ (with JetStream enabled)
- Network connectivity to NATS server

### For RabbitMQ Audit Trail
- RabbitMQ 3.10+ with AMQP support
- Queue configured for audit events

---

## Deployment Modes

Prysm supports multiple deployment modes depending on your use case:

| Mode | Use Case | Complexity | Scalability |
|------|----------|------------|-------------|
| **Standalone** | Single node, Prometheus-only | Low | Limited |
| **Distributed** | Multi-node with NATS | Medium | High |
| **Kubernetes** | Container orchestration | High | Very High |
| **Sidecar** | Automatic RGW integration | Medium | High |

---

## Standalone Deployment

### Building from Source

```bash
# Clone the repository
git clone https://github.com/cobaltcore-dev/prysm.git
cd prysm

# Build the binary
go build -o prysm ./cmd/main.go

# Move to system path
sudo mv prysm /usr/local/bin/
```

### Using Docker

```bash
# Pull the latest image
docker pull ghcr.io/cobaltcore-dev/prysm:latest

# Or build locally
docker build -t prysm:local .
```

### Running Standalone Producers

#### 1. Operations Log Producer (Prometheus Only)

```bash
# Basic configuration
prysm local-producer ops-log \
  --log-file /var/log/ceph/ceph-rgw-ops.json.log \
  --prometheus \
  --prometheus-port 8080 \
  --track-latency-per-method \
  --track-requests-per-tenant \
  -v info
```

#### 2. Disk Health Metrics

```bash
# Monitor all disks
prysm local-producer disk-health-metrics \
  --disks "*" \
  --interval 60 \
  --prometheus \
  --prometheus-port 8081 \
  --ceph-osd-base-path /var/lib/rook/rook-ceph \
  -v info
```

#### 3. Resource Usage Monitoring

```bash
prysm local-producer resource-usage \
  --prometheus \
  --prometheus-port 8082 \
  --interval 10 \
  -v info
```

### Systemd Service Configuration

Create a systemd service for persistent operation:

```ini
# /etc/systemd/system/prysm-ops-log.service
[Unit]
Description=Prysm Operations Log Producer
After=network.target

[Service]
Type=simple
User=ceph
Group=ceph
ExecStart=/usr/local/bin/prysm local-producer ops-log \
  --log-file /var/log/ceph/ceph-rgw-ops.json.log \
  --prometheus \
  --prometheus-port 8080 \
  --track-everything \
  -v info
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable prysm-ops-log
sudo systemctl start prysm-ops-log
sudo systemctl status prysm-ops-log
```

---

## Kubernetes Deployment

### Basic Deployment (Without Sidecar Injection)

#### 1. Create Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: prysm-monitoring
```

#### 2. Deploy NATS Server (Optional)

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nats
  namespace: prysm-monitoring
spec:
  serviceName: nats
  replicas: 1
  selector:
    matchLabels:
      app: nats
  template:
    metadata:
      labels:
        app: nats
    spec:
      containers:
      - name: nats
        image: nats:2.12-alpine
        ports:
        - containerPort: 4222
          name: client
        - containerPort: 8222
          name: monitoring
        args:
        - "-js"  # Enable JetStream
        - "-sd"  # Enable JetStream storage
        - "/data"
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: nats
  namespace: prysm-monitoring
spec:
  selector:
    app: nats
  ports:
  - port: 4222
    name: client
  - port: 8222
    name: monitoring
```

#### 3. Deploy Prysm Producer as DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: prysm-disk-health
  namespace: prysm-monitoring
spec:
  selector:
    matchLabels:
      app: prysm-disk-health
  template:
    metadata:
      labels:
        app: prysm-disk-health
    spec:
      hostNetwork: true
      hostPID: true
      containers:
      - name: prysm
        image: ghcr.io/cobaltcore-dev/prysm:v1.2.3
        args:
        - "local-producer"
        - "disk-health-metrics"
        - "--disks=*"
        - "--interval=60"
        - "--prometheus"
        - "--prometheus-port=8081"
        - "--nats-url=nats://nats.prysm-monitoring:4222"
        - "--nats-subject=osd.disk.health"
        - "-v=info"
        securityContext:
          privileged: true
        volumeMounts:
        - name: dev
          mountPath: /dev
        - name: sys
          mountPath: /sys
        ports:
        - containerPort: 8081
          name: metrics
      volumes:
      - name: dev
        hostPath:
          path: /dev
      - name: sys
        hostPath:
          path: /sys
```

#### 4. Create ServiceMonitor for Prometheus

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prysm-disk-health
  namespace: prysm-monitoring
spec:
  selector:
    matchLabels:
      app: prysm-disk-health
  endpoints:
  - port: metrics
    interval: 30s
```

---

## Distributed Deployment with NATS

This mode enables multiple producers and consumers across different nodes.

### Architecture Overview

```
Node A (Producer) ──┐
                    │
Node B (Producer) ──┼──► NATS Server ──► Consumer(s)
                    │
Node C (Producer) ──┘
```

### 1. Deploy NATS with Clustering

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nats-config
  namespace: prysm-monitoring
data:
  nats.conf: |
    port: 4222
    http: 8222

    jetstream {
      store_dir: /data
      max_memory_store: 1GB
      max_file_store: 10GB
    }

    cluster {
      name: prysm-nats-cluster
      port: 6222
      routes = [
        nats://nats-0.nats:6222
        nats://nats-1.nats:6222
        nats://nats-2.nats:6222
      ]
    }
```

### 2. Producer Configuration with NATS

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prysm-producer-config
  namespace: prysm-monitoring
data:
  config.yaml: |
    global:
      nats_url: "nats://nats.prysm-monitoring:4222"
      node_name: "${NODE_NAME}"
      instance_id: "${POD_NAME}"

    producers:
      - type: "ops_log"
        settings:
          nats_subject: "rgw.s3.ops"
          log_file: "/var/log/ceph/ops-log.log"
          prometheus: true
          prometheus_port: 8080
```

### 3. Deploy Consumer

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prysm-quota-consumer
  namespace: prysm-monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prysm-quota-consumer
  template:
    metadata:
      labels:
        app: prysm-quota-consumer
    spec:
      containers:
      - name: consumer
        image: ghcr.io/cobaltcore-dev/prysm:v1.2.3
        args:
        - "consumer"
        - "quota-usage"
        - "--nats-url=nats://nats.prysm-monitoring:4222"
        - "--nats-subject=rgw.quota.usage"
        - "--prometheus"
        - "--prometheus-port=8083"
        - "-v=info"
        ports:
        - containerPort: 8083
          name: metrics
```

---

## Kubernetes Sidecar Injection

This is the recommended deployment method for Ceph RadosGW monitoring.

### Prerequisites

1. Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

2. Verify cert-manager is running:
```bash
kubectl get pods -n cert-manager
```

### 1. Create Webhook Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: prysm-webhook
```

### 2. Deploy Certificate Resources

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: prysm-selfsigned-issuer
  namespace: prysm-webhook
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: prysm-webhook-cert
  namespace: prysm-webhook
spec:
  secretName: prysm-webhook-cert
  dnsNames:
    - prysm-webhook-service.prysm-webhook.svc
    - prysm-webhook-service.prysm-webhook.svc.cluster.local
  issuerRef:
    name: prysm-selfsigned-issuer
    kind: Issuer
```

### 3. Deploy Webhook Server

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prysm-webhook
  namespace: prysm-webhook
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prysm-webhook
  template:
    metadata:
      labels:
        app: prysm-webhook
    spec:
      containers:
      - name: webhook
        image: ghcr.io/cobaltcore-dev/prysm-wh:v1.2.3
        ports:
        - containerPort: 8443
          name: webhook
        env:
        - name: WEBHOOK_PORT
          value: "8443"
        - name: SIDECAR_IMAGE
          value: "ghcr.io/cobaltcore-dev/prysm:v1.2.3"
        volumeMounts:
        - name: certs
          mountPath: /certs
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8443
            scheme: HTTPS
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8443
            scheme: HTTPS
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: certs
        secret:
          secretName: prysm-webhook-cert
---
apiVersion: v1
kind: Service
metadata:
  name: prysm-webhook-service
  namespace: prysm-webhook
spec:
  selector:
    app: prysm-webhook
  ports:
  - port: 443
    targetPort: 8443
    name: webhook
```

### 4. Create MutatingWebhookConfiguration

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: prysm-sidecar-injector
  annotations:
    cert-manager.io/inject-ca-from: prysm-webhook/prysm-webhook-cert
webhooks:
  - name: prysm-sidecar.injector.webhook
    clientConfig:
      service:
        name: prysm-webhook-service
        namespace: prysm-webhook
        path: "/mutate"
    admissionReviewVersions: ["v1"]
    sideEffects: None
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["apps"]
        apiVersions: ["v1"]
        resources: ["deployments"]
    namespaceSelector:
      matchLabels:
        prysm-injection: enabled
```

### 5. Configure Sidecar Settings

Create a ConfigMap or Secret for sidecar configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prysm-sidecar-config
  namespace: rook-ceph
data:
  LOG_FILE_PATH: "/var/log/ceph/ops-log.log"
  MAX_LOG_FILE_SIZE: "10"
  PROMETHEUS_PORT: "9090"
  IGNORE_ANONYMOUS_REQUESTS: "true"
  TRACK_LATENCY_PER_METHOD: "true"
  TRACK_REQUESTS_PER_TENANT: "true"
  TRACK_ERRORS_PER_USER: "true"
  TRACK_BYTES_SENT_PER_BUCKET: "true"
```

For sensitive data (e.g., NATS credentials):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: prysm-sidecar-secret
  namespace: rook-ceph
type: Opaque
stringData:
  NATS_URL: "nats://user:password@nats.prysm-monitoring:4222"
  NATS_SUBJECT: "rgw.s3.ops"
```

### 6. Enable Injection on RGW

Modify your CephObjectStore to enable sidecar injection:

```yaml
apiVersion: ceph.rook.io/v1
kind: CephObjectStore
metadata:
  name: my-store
  namespace: rook-ceph
  annotations:
    prysm-sidecar/sidecar-env-configmap: "prysm-sidecar-config"
    prysm-sidecar/sidecar-env-secret: "prysm-sidecar-secret"
spec:
  gateway:
    port: 80
    instances: 2
    labels:
      prysm-sidecar: "yes"  # Enable injection
```

### 7. Verify Injection

```bash
# Check webhook is running
kubectl get pods -n prysm-webhook

# Check RGW pods have sidecar
kubectl get pods -n rook-ceph -l app=rook-ceph-rgw

# Verify sidecar container exists
kubectl describe pod <rgw-pod-name> -n rook-ceph | grep prysm-sidecar

# Check metrics endpoint
kubectl port-forward -n rook-ceph <rgw-pod-name> 9090:9090
curl http://localhost:9090/metrics
```

---

## Configuration Management

### Environment Variables

All configuration options can be set via environment variables:

```bash
# Global settings
export NODE_NAME="node-01"
export INSTANCE_ID="prysm-001"
export VERBOSITY="info"

# NATS settings
export NATS_URL="nats://localhost:4222"
export NATS_SUBJECT="rgw.s3.ops"

# Prometheus settings
export PROMETHEUS_PORT="8080"
export PROMETHEUS_INTERVAL="60"

# Feature flags
export TRACK_EVERYTHING="true"
export IGNORE_ANONYMOUS_REQUESTS="true"
```

### Configuration File

Use YAML configuration for complex setups:

```yaml
# config.yaml
global:
  nats_url: "nats://nats-jetstream.monitoring:4222"
  admin_url: "http://radosgw.rook-ceph:7480"
  access_key: "${RGW_ACCESS_KEY}"
  secret_key: "${RGW_SECRET_KEY}"
  node_name: "${NODE_NAME}"
  instance_id: "${INSTANCE_ID}"

producers:
  - type: "ops_log"
    settings:
      nats_subject: "rgw.s3.ops"
      log_file: "/var/log/ceph/ops-log.log"
      prometheus: true
      prometheus_port: 8080
      track_everything: true
      ignore_anonymous_requests: true

  - type: "disk_health_metrics"
    settings:
      nats_subject: "osd.disk.health"
      disks: ["*"]
      interval: 60
      prometheus: true
      prometheus_port: 8081
      ceph_osd_base_path: "/var/lib/rook/rook-ceph"
```

Load configuration:
```bash
prysm --config config.yaml local-producer ops-log
```

### Kubernetes ConfigMap Pattern

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prysm-config
  namespace: prysm-monitoring
data:
  config.yaml: |
    # Configuration content here
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prysm-producer
spec:
  template:
    spec:
      containers:
      - name: prysm
        volumeMounts:
        - name: config
          mountPath: /etc/prysm
        command:
        - prysm
        - --config
        - /etc/prysm/config.yaml
        - local-producer
        - ops-log
      volumes:
      - name: config
        configMap:
          name: prysm-config
```

---

## Production Considerations

### Resource Requirements

#### Minimal Configuration
```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

#### Heavy Workload (track-everything enabled)
```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

### High Availability

1. **Multiple Webhook Replicas**:
```yaml
spec:
  replicas: 3
```

2. **NATS Clustering**: Deploy 3+ NATS instances

3. **Consumer Scaling**: Use HPA for consumers
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: prysm-consumer-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: prysm-quota-consumer
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Security Hardening

1. **Non-root containers** (where possible):
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  capabilities:
    drop:
    - ALL
```

2. **Network Policies**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: prysm-webhook-policy
spec:
  podSelector:
    matchLabels:
      app: prysm-webhook
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 8443
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: rook-ceph
```

3. **RBAC Configuration**:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prysm-webhook
  namespace: prysm-webhook
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prysm-webhook-role
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prysm-webhook-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prysm-webhook-role
subjects:
- kind: ServiceAccount
  name: prysm-webhook
  namespace: prysm-webhook
```

### Monitoring Prysm Itself

Deploy ServiceMonitor for all components:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prysm-monitoring
  namespace: prysm-monitoring
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: prysm
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### Backup and Recovery

1. **NATS JetStream Snapshots**: Regular backups of stream state
2. **Configuration Backups**: Store configs in version control
3. **Disaster Recovery**: Document recovery procedures

---

## Troubleshooting

### Common Issues

#### 1. Sidecar Not Injected

**Symptoms**: RGW pods don't have prysm-sidecar container

**Debugging**:
```bash
# Check webhook is running
kubectl get pods -n prysm-webhook

# Check webhook logs
kubectl logs -n prysm-webhook deployment/prysm-webhook

# Verify webhook configuration
kubectl get mutatingwebhookconfigurations prysm-sidecar-injector -o yaml

# Check CA bundle is injected
kubectl get mutatingwebhookconfigurations prysm-sidecar-injector \
  -o jsonpath='{.webhooks[0].clientConfig.caBundle}' | base64 -d

# Check CephObjectStore labels
kubectl get cephobjectstore -n rook-ceph -o yaml
```

**Solutions**:
- Ensure `prysm-sidecar: "yes"` label is set in CephObjectStore
- Verify cert-manager injected CA bundle
- Check webhook service is accessible
- Review webhook pod logs for errors

#### 2. Metrics Not Appearing

**Symptoms**: Prometheus can't scrape metrics or metrics are empty

**Debugging**:
```bash
# Check if metrics endpoint is accessible
kubectl port-forward -n rook-ceph <pod-name> 9090:9090
curl http://localhost:9090/metrics

# Check prysm-sidecar logs
kubectl logs -n rook-ceph <pod-name> -c prysm-sidecar

# Verify log file exists and is being written
kubectl exec -n rook-ceph <pod-name> -c prysm-sidecar -- ls -la /var/log/ceph/
```

**Solutions**:
- Verify log file path is correct
- Check file permissions
- Ensure RadosGW is writing logs
- Verify metric tracking flags are enabled

#### 3. NATS Connection Failures

**Symptoms**: "connection refused" or timeout errors

**Debugging**:
```bash
# Check NATS is running
kubectl get pods -n prysm-monitoring -l app=nats

# Test NATS connectivity
kubectl run -it --rm nats-test --image=nats:2.12-alpine --restart=Never -- \
  nats pub -s nats://nats.prysm-monitoring:4222 test "hello"

# Check producer logs
kubectl logs <producer-pod> | grep -i nats
```

**Solutions**:
- Verify NATS service DNS resolution
- Check network policies
- Ensure NATS is listening on correct port
- Verify credentials if authentication is enabled

#### 4. High Memory Usage

**Symptoms**: OOMKilled pods or high memory consumption

**Debugging**:
```bash
# Check memory usage
kubectl top pod <pod-name>

# Review configuration
kubectl exec <pod-name> -- env | grep TRACK_
```

**Solutions**:
- Disable unnecessary metric tracking
- Use less granular metrics (per-tenant vs per-bucket)
- Increase memory limits
- Consider splitting producers

#### 5. Webhook Certificate Issues

**Symptoms**: Webhook calls fail with TLS errors

**Debugging**:
```bash
# Check certificate status
kubectl get certificate -n prysm-webhook

# Describe certificate
kubectl describe certificate prysm-webhook-cert -n prysm-webhook

# Check secret exists
kubectl get secret prysm-webhook-cert -n prysm-webhook
```

**Solutions**:
- Ensure cert-manager is running
- Verify certificate is Ready
- Recreate certificate if needed
- Check cert-manager logs

### Logging and Diagnostics

Enable debug logging:
```bash
# For standalone
prysm local-producer ops-log -v debug

# For Kubernetes (edit deployment)
args:
  - "local-producer"
  - "ops-log"
  - "-v=debug"
```

Collect diagnostics:
```bash
# Get all relevant logs
kubectl logs -n rook-ceph <rgw-pod> -c prysm-sidecar > prysm-sidecar.log
kubectl logs -n prysm-webhook deployment/prysm-webhook > webhook.log

# Get configuration
kubectl get cephobjectstore -n rook-ceph -o yaml > cephobjectstore.yaml
kubectl get mutatingwebhookconfigurations prysm-sidecar-injector -o yaml > webhook-config.yaml

# Get events
kubectl get events -n rook-ceph --sort-by='.lastTimestamp'
```

### Performance Tuning

1. **Optimize Metric Tracking**:
```yaml
# Minimal metrics
TRACK_LATENCY_PER_METHOD: "true"
TRACK_REQUESTS_PER_TENANT: "true"
TRACK_ERRORS_PER_USER: "true"

# vs. Maximum metrics (high overhead)
TRACK_EVERYTHING: "true"
```

2. **Adjust Collection Intervals**:
```yaml
# Disk health - less frequent
--interval=300  # 5 minutes

# Operations log - real-time
# (no interval, event-driven)
```

3. **NATS Performance**:
```yaml
# Enable JetStream for durability
jetstream:
  enabled: true
  max_memory_store: 1GB
  max_file_store: 10GB
```

---

## Next Steps

After successful deployment:

1. **Configure Prometheus** to scrape Prysm metrics
2. **Set up Grafana dashboards** for visualization
3. **Configure alerting rules** based on thresholds
4. **Review and optimize** metric collection based on actual usage
5. **Monitor resource consumption** and adjust limits

See [NEXT_STEPS.md](./NEXT_STEPS.md) for detailed guidance on post-deployment activities.
