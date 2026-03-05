# Prysm Design Comparison: NG vs NG-Small

**Quick Decision Guide:** Which design should we implement?

---

## TL;DR - Decision Matrix

| If you need... | Choose |
|----------------|--------|
| **Edge deployment, minimal footprint** | NG-Small ✅ |
| **Kubernetes sidecar (100+ pods)** | NG-Small ✅ |
| **Cost optimization (<64MB RAM per instance)** | NG-Small ✅ |
| **Simple monitoring needs** | NG-Small ✅ |
| **Raspberry Pi / IoT deployment** | NG-Small ✅ |
| **High Availability with state** | NG (Full) |
| **Complex stream processing** | NG (Full) |
| **Plugin system** | NG (Full) |
| **Enterprise features (RBAC, audit, etc.)** | NG (Full) |
| **Multi-region coordination** | NG (Full) |

---

## Side-by-Side Comparison

| Feature | Prysm v1 | Prysm-NG (Full) | **Prysm-NG-Small** |
|---------|----------|-----------------|-------------------|
| **Binary Size** | 20MB | 40MB | **<15MB ✅** |
| **Memory (Idle)** | 100MB | 200MB | **<20MB ✅** |
| **Memory (Active)** | 256MB | 512MB-2GB | **<50MB ✅** |
| **Startup Time** | 5s | 10s | **<1s ✅** |
| **Configuration** | ~100 lines | ~500 lines | **~50 lines ✅** |
| **Dependencies** | NATS (opt.) | NATS, PostgreSQL, etcd | **None ✅** |
| **External Services** | Optional | Required | **Optional ✅** |
| **Complexity** | Medium | High | **Low ✅** |
| **HA** | ❌ No | ✅ Yes (built-in) | ⚠️ Deploy multiple |
| **Persistence** | ❌ No | ✅ Yes (TimeSeries + State) | ❌ No (by design) |
| **State Storage** | ❌ No | ✅ PostgreSQL | ❌ No |
| **Leader Election** | ❌ No | ✅ Yes (etcd/K8s) | ❌ No |
| **Failover** | ❌ No | ✅ Automatic | ⚠️ Manual (LB) |
| **Plugin System** | ❌ No | ✅ SDK provided | ❌ No |
| **Stream Processing** | ❌ No | ✅ Full (windowing, joins) | ⚠️ Basic (transforms) |
| **mTLS** | ⚠️ Optional | ✅ Built-in | ⚠️ Optional |
| **RBAC** | ❌ No | ✅ Yes | ❌ No |
| **Vault Integration** | ❌ No | ✅ Yes | ❌ No |
| **OpenTelemetry** | ❌ No | ✅ Full | ⚠️ Metrics only |
| **GraphQL API** | ❌ No | ✅ Yes | ❌ No |
| **REST API** | ❌ No | ✅ Yes | ❌ No |
| **Hot Reload** | ❌ No | ✅ Via API | ✅ Via signal (USR1) |
| **Config Validation** | ⚠️ Basic | ✅ Comprehensive | ✅ Pre-start |
| **Test Coverage** | 6.7% | Target: 80% | Target: 85% |
| **Production Ready** | ❌ No | Timeline: 15 months | Timeline: 9 months ✅ |
| **Development Time** | Done | 12-15 months | **6-9 months ✅** |
| **Maintenance** | Medium | High | **Low ✅** |
| **Suitable For** | Dev/Test | Enterprise | **Edge/Scale ✅** |

---

## Resource Comparison

### Kubernetes Sidecar Example

**Prysm v1:**
```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

# Cost per pod: ~$12/month (AWS)
# 100 pods: ~$1,200/month
```

**Prysm-NG (Full):**
```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

# Cost per pod: ~$12/month (AWS)
# 100 pods: ~$1,200/month
# Plus: PostgreSQL, etcd, NATS
# Total: ~$1,500/month
```

**Prysm-NG-Small:**
```yaml
resources:
  requests:
    cpu: 50m        # 10x less CPU
    memory: 32Mi    # 16x less memory
  limits:
    cpu: 200m
    memory: 64Mi

# Cost per pod: ~$1.50/month (AWS)
# 100 pods: ~$150/month
# No external services required
# Total: ~$150/month ✅ 10x cheaper
```

### Real-World Scale Example

**Scenario:** 500 RGW pods across 3 regions

| Solution | CPU Cores | Memory | Monthly Cost (AWS) |
|----------|-----------|--------|-------------------|
| Prysm v1 | 250 cores | 250 GB | ~$6,000 |
| NG (Full) + Infra | 250 cores | 250 GB + DBs | ~$7,500 |
| **NG-Small** | **25 cores** | **25 GB** | **~$750 ✅** |

**Savings with NG-Small: $6,750/month ($81K/year)**

---

## Feature Comparison

### Configuration

**NG (Full) - Comprehensive but Complex:**
```yaml
# 500+ lines of configuration
apiVersion: prysm.io/v1
kind: Configuration
metadata:
  name: prysm-ng-production
  namespace: monitoring
  version: "2.0"

global:
  cluster_id: "prod-us-east-1"
  logging:
    level: info
    format: json
    output: stdout
    file:
      path: /var/log/prysm-ng/app.log
      max_size: 100
      max_age: 30
  metrics:
    enabled: true
    port: 9090
  tracing:
    enabled: true
    provider: opentelemetry
  # ... 450 more lines ...
```

**NG-Small - Minimal and Clear:**
```yaml
# ~50 lines for same functionality
sources:
  logs:
    type: file
    path: /var/log/ceph/ops-log.log

transforms:
  parse:
    type: remap
    inputs: [logs]
    source: |
      . = parse_json!(.message)
      .tenant = split(.user, "$")[0] ?? "default"

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [parse]
    address: 0.0.0.0:9090
```

### Error Handling

**NG (Full):**
```yaml
error_handling:
  default_strategy: graceful_degradation
  retry:
    enabled: true
    max_attempts: 3
    initial_backoff: 100ms
    max_backoff: 30s
    backoff_multiplier: 2.0
    jitter: true
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    success_threshold: 2
    timeout: 60s
  degraded_mode:
    enabled: true
    features_to_disable:
      - audit_trail
      - optional_metrics
```

**NG-Small:**
```yaml
# Built-in defaults, simple override
global:
  on_error: log_and_continue  # or drop, or retry
  retry:
    max_attempts: 3
    initial_delay: 100ms
```

---

## Architecture Comparison

### NG (Full) - Enterprise Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Configuration Layer                        │
│  etcd + Consul + LaunchDarkly + Config API                  │
└────────────────────────┬────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Control Plane                            │
│  Leader Election + Config Controller + Health Monitor       │
└────────────────────────┬────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Data Plane                              │
│  Producers → Stream Processing → Consumers                  │
└────────────────────────┬────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Storage Layer                             │
│  VictoriaMetrics + PostgreSQL + S3/MinIO                    │
└─────────────────────────────────────────────────────────────┘

Dependencies: NATS, PostgreSQL, etcd/Consul, VictoriaMetrics
```

### NG-Small - Minimal Architecture

```
┌────────────────────────────────────┐
│       Configuration (YAML)         │
└────────────────┬───────────────────┘
                 ▼
┌────────────────────────────────────┐
│        Pipeline Engine             │
│  Sources → Transforms → Sinks      │
│   (All in-memory ring buffer)      │
└────────────────┬───────────────────┘
                 ▼
         Optional Outputs:
         • Prometheus /metrics
         • NATS (if configured)
         • Console

Dependencies: None (all optional)
```

---

## Use Case Fit

### When to Choose NG (Full)

✅ **Enterprise Deployment**
- Need centralized management
- Multiple teams/clusters
- Compliance requirements (audit logging, RBAC)
- Budget for infrastructure

✅ **Complex Requirements**
- Advanced stream processing
- Multi-region coordination
- Plugin development
- Custom processors

✅ **HA Requirements**
- Cannot tolerate downtime
- Need automatic failover
- State persistence critical
- SLA commitments

✅ **Mature Organization**
- Dedicated ops team
- Existing infrastructure (PostgreSQL, etcd)
- Complex monitoring needs
- Custom integrations

### When to Choose NG-Small

✅ **Scale Deployment**
- 100+ monitoring instances
- Cost is a concern
- Minimal footprint required
- Sidecar pattern

✅ **Simple Requirements**
- Just need metrics/logs
- No complex processing
- Standard use cases
- Quick deployment

✅ **Resource Constrained**
- Edge deployments
- IoT devices
- Free-tier cloud
- Raspberry Pi

✅ **Fast Development**
- Need solution in 6-9 months
- Simple to maintain
- Low operational burden
- Easy to understand

---

## Migration Paths

### From Prysm v1

**To NG (Full):**
- More complex migration
- Need to set up infrastructure (PostgreSQL, etcd)
- Config conversion tool provided
- Parallel running for 4-8 weeks
- Timeline: 2-3 months

**To NG-Small:**
- Simple migration
- No new infrastructure
- Config is simpler (easier to convert)
- Can run in parallel immediately
- Timeline: 1-2 weeks ✅

### Between NG Variants

**From NG-Small → NG (Full):**
- Add infrastructure
- Expand configuration
- Enable advanced features
- Timeline: 1 month

**From NG (Full) → NG-Small:**
- Remove infrastructure dependencies
- Simplify configuration
- Lose HA/persistence
- Timeline: 2 weeks

---

## Cost Analysis (3 Years)

### Scenario: 200 RGW pods

**Prysm v1:**
- Infrastructure: $2,400/month
- Maintenance: 1 FTE @ $120K/year
- Total 3 years: $86,400 + $360K = **$446K**

**NG (Full):**
- Infrastructure: $3,000/month (includes DBs)
- Maintenance: 0.5 FTE @ $60K/year
- Total 3 years: $108,000 + $180K = **$288K**
- Savings vs v1: $158K

**NG-Small:**
- Infrastructure: $300/month
- Maintenance: 0.2 FTE @ $24K/year
- Total 3 years: $10,800 + $72K = **$83K**
- Savings vs v1: $363K ✅
- Savings vs NG-Full: $205K ✅

---

## Development Timeline Comparison

### NG (Full)

**Phase 1 (Months 1-3):** Foundation
- Configuration system
- Error handling
- HA architecture

**Phase 2 (Months 4-6):** Data Plane
- Refactored producers
- Stream processing
- Consumer groups

**Phase 3 (Months 7-9):** Operations
- OpenTelemetry
- Security
- Auto-scaling

**Phase 4 (Months 10-12):** Advanced
- Plugins
- ML features
- Cost optimization

**Phase 5 (Months 13-15):** GA
- Testing
- Documentation
- Release

**Total: 15 months**

### NG-Small

**Phase 1 (Months 1-2):** Core Engine
- Pipeline engine
- Basic transforms
- Prometheus sink

**Phase 2 (Months 3-4):** Components
- All sources
- All sinks
- All transforms

**Phase 3 (Months 5-6):** Ceph Features
- S3 parsing
- SMART normalization
- Enrichment

**Phase 4 (Months 7-8):** Polish
- Optimization
- Testing
- Documentation

**Phase 5 (Month 9):** GA
- Security audit
- Release

**Total: 9 months ✅**

---

## Recommendation

### Immediate Term (Next 6-12 months)

**Build NG-Small First:**

**Reasons:**
1. ✅ **Faster to market:** 9 months vs 15 months
2. ✅ **Lower risk:** Simpler implementation
3. ✅ **Addresses primary pain:** High resource usage in v1
4. ✅ **Solves scale problem:** 10x cost reduction
5. ✅ **Tests architecture:** Validates minimal approach

**Strategy:**
- Month 1-9: Develop NG-Small
- Month 6-9: Beta testing with real users
- Month 9: GA release
- Month 10-12: Gather feedback

### Long Term (12+ months)

**Evaluate NG (Full) based on feedback:**

**If users need:**
- HA features → Build NG-Full
- Complex processing → Build NG-Full
- Plugin system → Build NG-Full
- They're happy with NG-Small → Stop here ✅

**Phased Approach:**
1. **Now - Month 9:** NG-Small to GA
2. **Month 10-12:** Evaluate market need
3. **Month 13+:** NG-Full only if required

---

## Final Verdict

### Build Prysm-NG-Small

**Confidence:** High (90%)

**Rationale:**
1. Solves the primary problem (footprint) immediately
2. 10x cost reduction is compelling
3. Faster to market (9 vs 15 months)
4. Lower risk (simpler implementation)
5. Can always build NG-Full later if needed
6. Vector proves minimal approach works
7. Most users don't need HA/persistence

**Risk Mitigation:**
- Design NG-Small with path to NG-Full in mind
- Keep configuration compatible
- Document upgrade path
- Plan for optional features

**Success Metrics:**
- <15MB binary ✅
- <50MB RAM ✅
- <1s startup ✅
- 100K events/s ✅
- 9 months to GA ✅
- 10x cost reduction ✅

---

## Next Steps

1. ✅ Review NG-Small design
2. ✅ Get ops team feedback on config style
3. ✅ Approve minimal approach
4. → Build core pipeline prototype (Month 1)
5. → Beta release (Month 6)
6. → GA release (Month 9)
7. → Evaluate NG-Full need (Month 12)

---

**Recommendation:** Start with NG-Small, evaluate NG-Full in 12 months based on real-world feedback.

**Expected Outcome:** NG-Small will be sufficient for 80% of use cases. Build NG-Full only if enterprise features are truly needed.
