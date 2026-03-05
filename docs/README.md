# Prysm Documentation

Welcome to the Prysm documentation! This directory contains comprehensive guides to help you understand, deploy, and extend Prysm.

## 📚 Documentation Overview

### [INDEX.md](./INDEX.md) ⭐ **Start Here**
**Complete Documentation Navigation Hub**

Your guide to all Prysm documentation with learning paths, use case mapping, and quick navigation:
- Document inventory (10 comprehensive docs)
- Use case → document mapping
- Learning paths for different roles
- Maturity status and recommendations

**Start here if you want to**: Navigate the documentation efficiently or understand what's available.

---

### [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) ⚠️ **Critical Reading**
**Current State Assessment**

Brutally honest evaluation of Prysm v1:
- Overall maturity: 5.35/10 (early beta)
- Test coverage: 6.7% (industry standard: 70-90%)
- 48 log.Fatal() calls causing crashes
- Comparison with competitors
- Production readiness assessment
- Known limitations and gaps

**Start here if you want to**: Evaluate Prysm for production, understand current limitations, or make informed decisions.

---

### [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md) ⭐ **Design Decision Guide**
**Choosing the Right Implementation Approach**

Side-by-side comparison of three design options:
- **Prysm v1** (current): 20MB binary, 256MB RAM, early beta
- **Prysm-NG-Small** (recommended): <15MB binary, <50MB RAM, 6-9 months, 10x cost reduction
- **Prysm-NG** (enterprise): 40MB binary, 512MB-2GB RAM, 12-15 months, full HA/persistence

Includes decision matrix, cost analysis, migration paths, and clear recommendation.

**Start here if you want to**: Decide which design to implement or plan the future architecture.

---

### [PRYSM_NG_SMALL_DESIGN.md](./PRYSM_NG_SMALL_DESIGN.md) ⭐ **Recommended Design**
**Minimal Footprint Solution**

Vector-inspired minimal design for edge/scale deployments:
- Pipeline architecture: Sources → Transforms → Sinks
- <15MB binary, <50MB RAM, <1s startup
- Zero dependencies (all optional)
- Simple YAML configuration (~50 lines)
- Perfect for: Kubernetes sidecars, edge, IoT, cost-sensitive deployments
- Timeline: 6-9 months to production

**Start here if you want to**: Build a lightweight, cost-effective observability solution.

---

### [PRYSM_NG_DESIGN.md](./PRYSM_NG_DESIGN.md)
**Enterprise Design**

Complete enterprise redesign for complex requirements:
- Configuration-first architecture (500+ line YAML)
- High availability (active-passive, active-active)
- Data persistence (TimeSeries DB, state store)
- Plugin system and advanced stream processing
- Full security architecture (mTLS, RBAC, Vault)
- Timeline: 12-15 months to production

**Start here if you want to**: Plan enterprise deployment with HA, persistence, and complex integrations.

---

### [ARCHITECTURE.md](./ARCHITECTURE.md)
**Understanding Prysm's Design**

Learn about Prysm's four-layered architecture and how components interact:
- Architecture layers (Consumers, NATS, Producers)
- Component interactions and data flow
- Technology stack and design decisions
- Scalability considerations
- Security architecture
- Extension points for customization

**Start here if you want to**: Understand how Prysm works at a high level, contribute to the project, or design integrations.

---

### [CODE_EXPLAINED.md](./CODE_EXPLAINED.md)
**Deep Dive into the Codebase**

Detailed explanation of how the code works internally:
- Code structure and organization
- Command-line interface implementation
- Producer and consumer implementations
- Data processing pipelines
- Configuration management
- Messaging layer internals
- Metrics collection patterns
- Key design patterns used

**Start here if you want to**: Contribute code, debug issues, understand implementation details, or build custom producers/consumers.

---

### [DEPLOYMENT.md](./DEPLOYMENT.md)
**Deploying Prysm in Production**

Comprehensive deployment guide covering all deployment modes:
- Prerequisites and system requirements
- Standalone deployment (systemd, Docker)
- Kubernetes deployment (DaemonSet, Deployment)
- Distributed deployment with NATS
- Kubernetes sidecar injection (automatic RGW monitoring)
- Configuration management
- Production considerations (HA, security, monitoring)
- Troubleshooting common issues

**Start here if you want to**: Deploy Prysm in your environment, set up monitoring for Ceph/RadosGW, or migrate to production.

---

### [NEXT_STEPS.md](./NEXT_STEPS.md)
**Post-Deployment and Future Roadmap**

What to do after deploying Prysm:
- Post-deployment verification
- Monitoring and observability setup (Prometheus, Grafana, Alerting)
- Integration with existing systems (OpenStack, ServiceNow, Slack)
- Performance optimization
- Security hardening
- Development and contribution guide
- Future enhancements roadmap
- Production readiness checklist

**Start here if you want to**: Configure monitoring dashboards, optimize performance, integrate with existing tools, or plan future enhancements.

### [QUICK_REFERENCE.md](./QUICK_REFERENCE.md)
**Command Cheat Sheet**

Quick reference for day-to-day operations:
- Common commands for all producers
- Prometheus queries
- Alert rules
- Environment variables
- Troubleshooting commands

**Start here if you want to**: Find commands quickly during daily operations.

---

## 🚀 Quick Start Paths

### I want to understand Prysm
1. Read the [main README](../README.md)
2. Review [INDEX.md](./INDEX.md) for navigation
3. Check [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) for current state
4. Browse [ARCHITECTURE.md](./ARCHITECTURE.md) for design details

### I want to evaluate Prysm
1. **CRITICAL**: Read [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) first
2. Review [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md) for future options
3. Check suitable use cases and limitations
4. Review [DEPLOYMENT.md](./DEPLOYMENT.md) if suitable

### I want to deploy Prysm
1. Check [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) to understand limitations
2. Verify prerequisites in [DEPLOYMENT.md](./DEPLOYMENT.md#prerequisites)
2. Choose your deployment mode
3. Follow the relevant deployment section
4. Verify installation using [DEPLOYMENT.md](./DEPLOYMENT.md#verify-installation)
5. Configure monitoring with [NEXT_STEPS.md](./NEXT_STEPS.md#monitoring-and-observability-setup)

### I want to contribute to Prysm
1. Read [INDEX.md](./INDEX.md) for documentation navigation
2. Review [CODE_EXPLAINED.md](./CODE_EXPLAINED.md) for implementation details
3. Check [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) to understand current issues
4. Review [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md) for future direction
5. See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines

### I want to plan the future of Prysm
1. **START HERE**: [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md)
2. Review recommended approach: [PRYSM_NG_SMALL_DESIGN.md](./PRYSM_NG_SMALL_DESIGN.md)
3. Check enterprise option if needed: [PRYSM_NG_DESIGN.md](./PRYSM_NG_DESIGN.md)
4. Understand current gaps: [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md)

### I want to use Prysm in production
**⚠️ IMPORTANT**: Prysm v1 is NOT production-ready for mission-critical use.
1. Read [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) to understand limitations
2. Review suitable use cases (dev/test, small deployments only)
3. If proceeding: Complete deployment using [DEPLOYMENT.md](./DEPLOYMENT.md)
4. Follow [DEPLOYMENT.md](./DEPLOYMENT.md#production-considerations)
5. Set up monitoring with [NEXT_STEPS.md](./NEXT_STEPS.md#monitoring-and-observability-setup)
6. For production-ready solution: Review [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md)

---

## 📋 Component-Specific Documentation

### Producers

#### Operations Log Producer
- **Main docs**: [pkg/producers/opslog/README.md](../pkg/producers/opslog/README.md)
- **Purpose**: Process RadosGW S3 operation logs
- **Use cases**: Request tracking, latency monitoring, audit trail
- **Deployment**: Sidecar injection or standalone

#### Disk Health Metrics
- **Main docs**: [pkg/producers/diskhealthmetrics/README.md](../pkg/producers/diskhealthmetrics/README.md)
- **Purpose**: Monitor disk health using SMART attributes
- **Use cases**: Predictive failure detection, capacity planning
- **Deployment**: DaemonSet on storage nodes

#### RadosGW Usage Exporter
- **Main docs**: [pkg/producers/radosgwusage/README.md](../pkg/producers/radosgwusage/README.md)
- **Purpose**: Export usage statistics from RadosGW Admin API
- **Use cases**: Usage reporting, billing, capacity tracking
- **Deployment**: Remote producer (outside cluster)

#### Quota Usage Monitor
- **Main docs**: [pkg/producers/quotausagemonitor/README.md](../pkg/producers/quotausagemonitor/README.md)
- **Purpose**: Monitor quota usage per user/bucket
- **Use cases**: Quota alerts, usage trends
- **Deployment**: Remote producer

### Consumers

#### Quota Usage Consumer
- **Main docs**: [pkg/consumer/quotausageconsumer/README.md](../pkg/consumer/quotausageconsumer/README.md)
- **Purpose**: Process and analyze quota usage events
- **Use cases**: Threshold alerts, usage analytics
- **Deployment**: Kubernetes Deployment

### Kubernetes Integration

#### Mutating Webhook for Sidecar Injection
- **Main docs**: [ops-log-k8s-mutating-wh/README.md](../ops-log-k8s-mutating-wh/README.md)
- **Purpose**: Automatically inject Prysm sidecar into RGW pods
- **Use cases**: Zero-config RGW monitoring
- **Deployment**: Webhook server with cert-manager

---

## 🏗️ Architecture Diagrams

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                        Consumers                             │
│   (Analyze data, generate alerts, store metrics)            │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    NATS Message Bus                          │
│         (Low-latency, high-throughput messaging)             │
└────────┬──────────────────────────────────────────┬─────────┘
         │                                          │
┌────────▼──────────────┐              ┌───────────▼──────────┐
│  Remote Producers     │              │  Nearby Producers    │
│  (API-based)          │              │  (Local access)      │
│  - Bucket Notify      │              │  - Ops Log           │
│  - Quota Monitor      │              │  - Disk Health       │
│  - RGW Usage          │              │  - Kernel Metrics    │
└───────────────────────┘              └──────────────────────┘
```

### Kubernetes Deployment
```
┌──────────────────────────────────────────────────────────┐
│ Kubernetes Cluster                                       │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ rook-ceph namespace                            │    │
│  │  ┌──────────────────────────────────────────┐  │    │
│  │  │ RGW Pod                                  │  │    │
│  │  │  ┌────────────┐    ┌──────────────────┐ │  │    │
│  │  │  │ RadosGW    │    │ Prysm Sidecar    │ │  │    │
│  │  │  │ Container  │───▶│ (auto-injected)  │ │  │    │
│  │  │  └────────────┘    └────────┬─────────┘ │  │    │
│  │  │                              │ :9090     │  │    │
│  │  └──────────────────────────────┼───────────┘  │    │
│  └─────────────────────────────────┼──────────────┘    │
│                                    │                    │
│  ┌─────────────────────────────────▼──────────────┐    │
│  │ monitoring namespace                           │    │
│  │  ┌──────────────┐   ┌─────────────────────┐   │    │
│  │  │ Prometheus   │──▶│ Grafana Dashboards  │   │    │
│  │  └──────────────┘   └─────────────────────┘   │    │
│  └────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
```

---

## 🔧 Configuration Examples

### Minimal Configuration (Low Overhead)
```yaml
producers:
  - type: "ops_log"
    settings:
      log_file: "/var/log/ceph/ops-log.log"
      prometheus: true
      prometheus_port: 8080
      track_latency_per_method: true
      track_requests_per_tenant: true
      track_errors_per_user: true
```

### Comprehensive Configuration (Maximum Visibility)
```yaml
global:
  nats_url: "nats://nats.monitoring:4222"

producers:
  - type: "ops_log"
    settings:
      log_file: "/var/log/ceph/ops-log.log"
      prometheus: true
      prometheus_port: 8080
      use_nats: true
      nats_subject: "rgw.s3.ops"
      track_everything: true
      audit_enabled: true
      audit_rabbitmq_url: "amqp://user:pass@rabbitmq:5672"

  - type: "disk_health_metrics"
    settings:
      disks: ["*"]
      interval: 60
      prometheus: true
      prometheus_port: 8081
      ceph_osd_base_path: "/var/lib/rook/rook-ceph"
```

---

## 📊 Metrics Reference

### RadosGW Operations Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `radosgw_total_requests` | Counter | Total requests with full detail |
| `radosgw_requests_duration` | Histogram | Request latency in seconds |
| `radosgw_bytes_sent` | Counter | Total bytes sent |
| `radosgw_errors_detailed` | Counter | Error counts by type |
| `radosgw_timeout_errors` | Counter | Timeout errors (OSD issues) |

### Disk Health Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `disk_temperature_celsius` | Gauge | Current disk temperature |
| `disk_reallocated_sectors` | Gauge | Reallocated sector count |
| `ssd_life_used_percentage` | Gauge | SSD wear level |
| `smart_attributes` | Gauge | Raw SMART attributes |
| `disk_info` | Info | Static disk information |

See individual producer documentation for complete metric lists.

---

## 🐛 Troubleshooting

### Common Issues

| Issue | Documentation | Quick Fix |
|-------|---------------|-----------|
| Sidecar not injected | [DEPLOYMENT.md](./DEPLOYMENT.md#1-sidecar-not-injected) | Check labels and webhook logs |
| Metrics not appearing | [DEPLOYMENT.md](./DEPLOYMENT.md#2-metrics-not-appearing) | Verify log file path and permissions |
| NATS connection failures | [DEPLOYMENT.md](./DEPLOYMENT.md#3-nats-connection-failures) | Check network policies and DNS |
| High memory usage | [DEPLOYMENT.md](./DEPLOYMENT.md#4-high-memory-usage) | Disable unnecessary metric tracking |
| Certificate issues | [DEPLOYMENT.md](./DEPLOYMENT.md#5-webhook-certificate-issues) | Verify cert-manager setup |

---

## 🤝 Contributing

We welcome contributions! Please see:
- [INDEX.md](./INDEX.md) - Documentation navigation and learning paths
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [CODE_EXPLAINED.md](./CODE_EXPLAINED.md) - Code walkthrough
- [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) - Known issues to address
- [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md) - Future direction
- [ARCHITECTURE.md](./ARCHITECTURE.md#extension-points) - Extension points

---

## 📝 Additional Resources

### Documentation Suite
- [INDEX.md](./INDEX.md) - Complete documentation navigation hub
- [HONEST_ANALYSIS.md](./HONEST_ANALYSIS.md) - Current state assessment (5.35/10)
- [DESIGN_COMPARISON.md](./DESIGN_COMPARISON.md) - Design decision guide
- [PRYSM_NG_SMALL_DESIGN.md](./PRYSM_NG_SMALL_DESIGN.md) - Recommended minimal design
- [PRYSM_NG_DESIGN.md](./PRYSM_NG_DESIGN.md) - Enterprise design option

### External Documentation
- [Ceph Documentation](https://docs.ceph.com/)
- [RadosGW Admin API](https://docs.ceph.com/en/latest/radosgw/adminops/)
- [NATS Documentation](https://docs.nats.io/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Rook Documentation](https://rook.io/docs/)

### Related Projects
- [Rook-Ceph](https://github.com/rook/rook) - Ceph operator for Kubernetes
- [NATS](https://github.com/nats-io/nats-server) - Messaging system
- [cert-manager](https://github.com/cert-manager/cert-manager) - Certificate management

---

## 📜 License

Copyright 2025 SAP SE or an SAP affiliate company and prysm contributors.

Licensed under the Apache License, Version 2.0. See [LICENSE](../LICENSE) for details.

---

## 📧 Getting Help

- **Issues**: [GitHub Issues](https://github.com/cobaltcore-dev/prysm/issues)
- **Discussions**: [GitHub Discussions](https://github.com/cobaltcore-dev/prysm/discussions)
- **Security**: See [SECURITY.md](https://github.com/cobaltcore-dev/prysm/security/policy)

---

**Note**: Prysm is under active development. Documentation is continuously updated. If you find errors or have suggestions, please open an issue or submit a pull request.
