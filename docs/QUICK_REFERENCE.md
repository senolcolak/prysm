# Prysm Quick Reference

## 🚀 Quick Commands

### Local Producers

#### Operations Log (S3/RadosGW)
```bash
# Minimal (Prometheus only)
prysm local-producer ops-log \
  --log-file /var/log/ceph/ops-log.log \
  --prometheus --prometheus-port 8080

# Recommended
prysm local-producer ops-log \
  --log-file /var/log/ceph/ops-log.log \
  --prometheus --prometheus-port 8080 \
  --track-latency-per-method \
  --track-requests-per-tenant \
  --track-errors-per-user

# Full monitoring
prysm local-producer ops-log \
  --log-file /var/log/ceph/ops-log.log \
  --prometheus --prometheus-port 8080 \
  --track-everything
```

#### Disk Health
```bash
# All disks
prysm local-producer disk-health-metrics \
  --disks "*" \
  --interval 60 \
  --prometheus --prometheus-port 8081

# Specific disks with Ceph integration
prysm local-producer disk-health-metrics \
  --disks "/dev/sda,/dev/sdb" \
  --interval 60 \
  --prometheus --prometheus-port 8081 \
  --ceph-osd-base-path /var/lib/rook/rook-ceph
```

#### Resource Usage
```bash
prysm local-producer resource-usage \
  --prometheus --prometheus-port 8082 \
  --interval 10
```

### Remote Producers

#### RadosGW Usage
```bash
prysm remote-producer radosgw-usage \
  --admin-url http://radosgw:7480 \
  --access-key <key> \
  --secret-key <secret> \
  --prometheus --prometheus-port 8083
```

#### Quota Usage Monitor
```bash
prysm remote-producer quota-usage-monitor \
  --admin-url http://radosgw:7480 \
  --access-key <key> \
  --secret-key <secret> \
  --nats-url nats://nats:4222 \
  --nats-subject rgw.quota.usage
```

### Consumers

#### Quota Usage Consumer
```bash
prysm consumer quota-usage \
  --nats-url nats://nats:4222 \
  --nats-subject rgw.quota.usage \
  --prometheus --prometheus-port 8084
```

---

## 📊 Common Prometheus Queries

### Request Metrics
```promql
# Total requests per second
rate(radosgw_total_requests[5m])

# Requests by tenant
sum(rate(radosgw_total_requests_per_tenant[5m])) by (tenant)

# Requests by method
sum(rate(radosgw_requests_by_method_global[5m])) by (method)
```

### Latency Metrics
```promql
# 95th percentile latency
histogram_quantile(0.95, rate(radosgw_requests_duration_bucket[5m]))

# Latency by method
histogram_quantile(0.95,
  rate(radosgw_requests_duration_per_method_bucket[5m])
) by (method)

# Average latency
rate(radosgw_requests_duration_sum[5m]) /
rate(radosgw_requests_duration_count[5m])
```

### Error Metrics
```promql
# Error rate
rate(radosgw_errors_detailed[5m])

# Error rate percentage
(rate(radosgw_errors_detailed[5m]) /
 rate(radosgw_total_requests[5m])) * 100

# Timeout errors (OSD issues)
rate(radosgw_timeout_errors[5m])
```

### Disk Health
```promql
# Disk temperature
disk_temperature_celsius

# High temperature alert
disk_temperature_celsius > 50

# Reallocated sectors
disk_reallocated_sectors

# SSD life used
ssd_life_used_percentage

# NVMe critical warnings
smart_attributes{attribute="critical_warning"} > 0
```

### Bytes Transferred
```promql
# Total bytes sent per second
rate(radosgw_bytes_sent[5m])

# Bandwidth by bucket
sum(rate(radosgw_bytes_sent_per_bucket[5m])) by (bucket)

# Total traffic (sent + received)
rate(radosgw_bytes_sent[5m]) + rate(radosgw_bytes_received[5m])
```

---

## 🎯 Alert Rules

### Operations
```yaml
# High error rate
- alert: HighErrorRate
  expr: |
    (rate(radosgw_errors_detailed[5m]) /
     rate(radosgw_total_requests[5m])) > 0.05
  for: 5m
  labels:
    severity: warning

# High latency
- alert: HighLatency
  expr: |
    histogram_quantile(0.95,
      rate(radosgw_requests_duration_bucket[5m])
    ) > 5
  for: 10m
  labels:
    severity: warning

# Timeout errors
- alert: TimeoutErrors
  expr: rate(radosgw_timeout_errors[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
```

### Disk Health
```yaml
# High temperature
- alert: DiskHighTemperature
  expr: disk_temperature_celsius > 60
  for: 5m
  labels:
    severity: warning

# Reallocated sectors
- alert: DiskReallocationWarning
  expr: disk_reallocated_sectors > 10
  for: 1m
  labels:
    severity: warning

# SSD wear
- alert: SSDWearWarning
  expr: ssd_life_used_percentage > 80
  for: 5m
  labels:
    severity: warning

# NVMe critical
- alert: NVMECriticalWarning
  expr: smart_attributes{attribute="critical_warning"} > 0
  for: 1m
  labels:
    severity: critical
```

---

## 🔧 Environment Variables

### Common
```bash
# Logging
VERBOSITY=info                    # debug, info, warn, error

# Node identification
NODE_NAME=node-01
INSTANCE_ID=prysm-001

# NATS
NATS_URL=nats://nats:4222
NATS_SUBJECT=rgw.s3.ops

# Prometheus
PROMETHEUS_PORT=8080
PROMETHEUS_INTERVAL=60
```

### Operations Log
```bash
# Input
LOG_FILE_PATH=/var/log/ceph/ops-log.log

# Features
TRACK_EVERYTHING=true
IGNORE_ANONYMOUS_REQUESTS=true
TRUNCATE_LOG_ON_START=false

# Specific tracking
TRACK_LATENCY_PER_METHOD=true
TRACK_REQUESTS_PER_TENANT=true
TRACK_ERRORS_PER_USER=true
TRACK_BYTES_SENT_PER_BUCKET=true

# Audit trail
AUDIT_ENABLED=true
AUDIT_RABBITMQ_URL=amqp://user:pass@rabbitmq:5672
AUDIT_QUEUE_NAME=keystone.notifications.info
```

### Disk Health
```bash
# Configuration
DISKS=/dev/sda,/dev/sdb
INTERVAL=60

# Ceph integration
CEPH_OSD_BASE_PATH=/var/lib/rook/rook-ceph

# Thresholds
GROWN_DEFECTS_THRESHOLD=10
PENDING_SECTORS_THRESHOLD=3
REALLOCATED_SECTORS_THRESHOLD=10
LIFETIME_USED_THRESHOLD=80
```

---

## 🐳 Docker Quick Start

### Standalone
```bash
docker run -d \
  --name prysm-ops-log \
  -v /var/log/ceph:/var/log/ceph:ro \
  -p 8080:8080 \
  ghcr.io/cobaltcore-dev/prysm:latest \
  local-producer ops-log \
  --log-file /var/log/ceph/ops-log.log \
  --prometheus --prometheus-port 8080
```

### With NATS
```bash
docker run -d \
  --name prysm-ops-log \
  -v /var/log/ceph:/var/log/ceph:ro \
  -p 8080:8080 \
  --network monitoring \
  ghcr.io/cobaltcore-dev/prysm:latest \
  local-producer ops-log \
  --log-file /var/log/ceph/ops-log.log \
  --prometheus --prometheus-port 8080 \
  --nats-url nats://nats:4222 \
  --nats-subject rgw.s3.ops
```

---

## ☸️ Kubernetes Quick Start

### Quick Deploy with Sidecar Injection

```bash
# 1. Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# 2. Deploy webhook
kubectl apply -f ops-log-k8s-mutating-wh/manifest-examples/

# 3. Enable injection on CephObjectStore
kubectl patch cephobjectstore my-store -n rook-ceph --type merge -p '
spec:
  gateway:
    labels:
      prysm-sidecar: "yes"
'

# 4. Restart RGW pods
kubectl rollout restart deployment -n rook-ceph -l app=rook-ceph-rgw
```

---

## 🔍 Troubleshooting Commands

### Check Status
```bash
# Systemd
systemctl status prysm-ops-log
journalctl -u prysm-ops-log -f

# Kubernetes
kubectl get pods -n prysm-monitoring
kubectl logs -n rook-ceph <rgw-pod> -c prysm-sidecar
kubectl describe pod -n rook-ceph <rgw-pod>

# Docker
docker logs prysm-ops-log -f
docker stats prysm-ops-log
```

### Test Metrics
```bash
# Local
curl http://localhost:8080/metrics

# Kubernetes
kubectl port-forward -n rook-ceph <pod> 9090:9090
curl http://localhost:9090/metrics

# Check specific metric
curl -s http://localhost:8080/metrics | grep radosgw_total_requests
```

### Test NATS
```bash
# Subscribe
nats sub -s nats://nats:4222 "rgw.s3.ops"

# Publish test
nats pub -s nats://nats:4222 "rgw.s3.ops" "test"

# Check connection
nats account info -s nats://nats:4222
```

### Webhook Issues
```bash
# Check webhook
kubectl get mutatingwebhookconfigurations prysm-sidecar-injector

# Check certificate
kubectl get certificate -n prysm-webhook
kubectl describe certificate prysm-webhook-cert -n prysm-webhook

# Check webhook logs
kubectl logs -n prysm-webhook deployment/prysm-webhook

# Verify CA injection
kubectl get mutatingwebhookconfigurations prysm-sidecar-injector \
  -o jsonpath='{.webhooks[0].clientConfig.caBundle}' | base64 -d
```

---

## 📈 Performance Tuning

### Resource Limits

**Light workload** (< 1000 req/sec):
```yaml
resources:
  requests: {cpu: 100m, memory: 256Mi}
  limits: {cpu: 500m, memory: 512Mi}
```

**Medium workload** (1000-10000 req/sec):
```yaml
resources:
  requests: {cpu: 500m, memory: 512Mi}
  limits: {cpu: 2000m, memory: 2Gi}
```

**Heavy workload** (> 10000 req/sec):
```yaml
resources:
  requests: {cpu: 1000m, memory: 1Gi}
  limits: {cpu: 4000m, memory: 4Gi}
```

### Metric Optimization

**Minimal** (lowest overhead):
```bash
--track-latency-per-method
--track-requests-per-tenant
--track-errors-per-user
```

**Balanced** (recommended):
```bash
--track-latency-detailed
--track-latency-per-method
--track-requests-per-user
--track-requests-per-bucket
--track-errors-per-user
```

**Maximum** (high overhead):
```bash
--track-everything
```

---

## 🔐 Security Checklist

- [ ] Use specific image tags (not `:latest`)
- [ ] Enable TLS for NATS
- [ ] Use Kubernetes Secrets for credentials
- [ ] Apply least-privilege RBAC
- [ ] Enable Pod Security Standards
- [ ] Set resource limits
- [ ] Use Network Policies
- [ ] Rotate credentials regularly
- [ ] Enable audit logging
- [ ] Use read-only root filesystem where possible

---

## 📚 Quick Links

- **Documentation**: [docs/README.md](./README.md)
- **Architecture**: [docs/ARCHITECTURE.md](./ARCHITECTURE.md)
- **Deployment**: [docs/DEPLOYMENT.md](./DEPLOYMENT.md)
- **Code Walkthrough**: [docs/CODE_EXPLAINED.md](./CODE_EXPLAINED.md)
- **Next Steps**: [docs/NEXT_STEPS.md](./NEXT_STEPS.md)
- **Issues**: https://github.com/cobaltcore-dev/prysm/issues
- **Discussions**: https://github.com/cobaltcore-dev/prysm/discussions

---

**Version**: Compatible with Prysm v1.0+
**Last Updated**: 2026-03-04
