# Prysm Architecture Documentation

## Overview

Prysm is a comprehensive observability CLI tool designed for Ceph and RadosGW (Rados Gateway) monitoring. It provides real-time monitoring, data collection, and analysis through a multi-layered architecture that enables flexible and scalable observability across diverse storage environments.

## Architecture Layers

Prysm implements a four-layered architecture that separates concerns and enables horizontal scalability:

```
┌─────────────────────────────────────────────────────────────┐
│                        Consumers                             │
│  (Process, analyze, alert, store, and visualize data)       │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    NATS Message Bus                          │
│  (Real-time messaging backbone with low-latency routing)    │
└────────┬──────────────────────────────────────────┬─────────┘
         │                                          │
┌────────▼──────────────┐              ┌───────────▼──────────┐
│  Remote Producers     │              │  Nearby Producers    │
│  (API-based)          │              │  (Direct access)     │
└───────────────────────┘              └──────────────────────┘
```

### 1. Consumers Layer

**Purpose**: Process and analyze data from various sources, generate alerts, and provide insights.

**Key Components**:
- **Quota Usage Consumer**: Monitors and analyzes quota usage across tenants and users
- Alert generation based on configurable thresholds
- Log storage and analysis for troubleshooting
- Real-time metrics visualization

**Data Flow**:
```
NATS Subject → Consumer → Processing → Output (Prometheus/Alerts/Logs)
```

**Responsibilities**:
- Subscribe to NATS subjects for data ingestion
- Apply business logic and thresholds
- Generate alerts for anomalies
- Export metrics to monitoring systems
- Maintain compliance records

### 2. NATS Message Bus

**Purpose**: Provides the messaging backbone for the entire system.

**Key Features**:
- **Low Latency**: Sub-millisecond message routing
- **High Throughput**: Handles millions of messages per second
- **Reliability**: Built-in message persistence with JetStream
- **Scalability**: Horizontal scaling support
- **Subject-Based Routing**: Flexible topic-based message distribution

**Architecture**:
```
Producer → NATS Subject (e.g., "rgw.s3.ops") → Multiple Consumers
                                             → Consumer A
                                             → Consumer B
                                             → Consumer N
```

**Key Subjects**:
- `rgw.s3.ops` - Raw S3 operation logs
- `rgw.s3.ops.aggregated.metrics` - Aggregated metrics
- `rgw.buckets.notify` - Bucket notification events
- `osd.disk.health` - Disk health metrics
- `osd.kernel.metrics` - Kernel-level metrics
- `osd.resource.usage` - Resource usage data

### 3. Remote Producers

**Purpose**: Collect metrics and logs from external APIs or interfaces without direct system access.

**Components**:

#### a. Bucket Notifications Producer
- Receives S3 bucket event notifications
- Publishes to NATS for downstream processing
- Supports filtering and routing

#### b. Quota Usage Monitor
- Polls RadosGW Admin API for quota information
- Tracks quota consumption per user and bucket
- Publishes usage metrics to NATS

#### c. RadosGW Usage Exporter
- Collects usage statistics from RadosGW Admin API
- Aggregates data at user and bucket levels
- Provides Prometheus-compatible metrics
- Reconciles state using NATS KV store

**Characteristics**:
- Deployed outside the storage cluster
- API-based data collection
- Network-accessible from monitoring infrastructure
- Minimal configuration required

### 4. Nearby Producers

**Purpose**: Collect data directly from systems with local access to logs, metrics, and hardware.

**Components**:

#### a. Operations Log Producer (ops-log)
**Most comprehensive producer** - Processes Ceph RadosGW S3 operation logs

**Features**:
- Real-time log parsing and processing
- Multiple metric aggregation levels
- RabbitMQ audit trail (CADF format)
- Log rotation management
- Configurable metric tracking

**Metric Categories**:
- Request counters (detailed, per-user, per-bucket, per-tenant)
- Method-based tracking (GET, PUT, DELETE, etc.)
- Operation-based tracking (S3 operations)
- Status-based tracking (HTTP status codes)
- Bytes transferred (sent/received)
- Error tracking with categorization
- Latency histograms
- IP-based metrics

**Output Options**:
- NATS subjects (raw + aggregated)
- Prometheus metrics endpoint
- RabbitMQ audit queue (Keystone-compatible)
- Console output

#### b. Disk Health Metrics Producer
**Hardware monitoring** - SMART attribute collection and normalization

**Features**:
- SMART attribute normalization across vendors
- NVMe-specific metrics support
- Ceph OSD integration (device-to-OSD mapping)
- LVM logical volume resolution
- Critical warning detection

**Metrics**:
- Temperature monitoring
- Reallocated sectors
- Pending sectors
- Power-on hours
- SSD life used percentage
- Error counts
- NVMe critical warnings
- Device information

#### c. Kernel Metrics Producer
**System-level monitoring** - Kernel metrics collection

**Features**:
- Kernel-level statistics
- Network statistics
- System resource monitoring

#### d. Resource Usage Producer
**Resource monitoring** - CPU, memory, and system resources

**Features**:
- CPU utilization
- Memory usage
- I/O statistics
- Resource trends

**Characteristics**:
- Deployed as sidecars or DaemonSets
- Direct file system access
- Local hardware access (SMART data)
- Low latency data collection
- Higher data fidelity

## Data Flow Architecture

### End-to-End Data Flow Example: S3 Operation Logging

```
┌──────────────┐
│ Ceph RadosGW │ writes logs
└──────┬───────┘
       │
       ▼
┌─────────────────────────────────┐
│ ops-log.log (JSON log file)     │
└──────────┬──────────────────────┘
           │
           ▼ (fsnotify watches)
┌──────────────────────────────────────────────────┐
│ Prysm ops-log Producer (Nearby Producer)         │
│ ┌──────────────────────────────────────────────┐ │
│ │ 1. Parse log entry                           │ │
│ │ 2. Extract metrics                           │ │
│ │ 3. Normalize data                            │ │
│ │ 4. Store in dedicated metric maps            │ │
│ └──────────────────────────────────────────────┘ │
└──┬───────────────────────────┬───────────────┬───┘
   │                           │               │
   ▼                           ▼               ▼
┌──────────┐          ┌─────────────┐   ┌──────────────┐
│   NATS   │          │ Prometheus  │   │  RabbitMQ    │
│ Subjects │          │   Metrics   │   │ (CADF Audit) │
└──────────┘          └─────────────┘   └──────────────┘
```

## Component Interactions

### Producer-NATS-Consumer Pattern

```
┌─────────────────┐
│ Producer        │
│ ┌─────────────┐ │
│ │ Data Source │ │
│ └──────┬──────┘ │
│        │        │
│ ┌──────▼──────┐ │
│ │ Processing  │ │
│ └──────┬──────┘ │
│        │        │
│ ┌──────▼──────┐ │
│ │ NATS Pub    │ │
│ └──────┬──────┘ │
└────────┼────────┘
         │
         ▼
┌────────────────────────────┐
│ NATS Server                │
│ Subject: rgw.s3.ops        │
└────────┬───────────────────┘
         │
         ├──────────┬─────────────┐
         ▼          ▼             ▼
┌────────────┐ ┌──────────┐ ┌─────────┐
│Consumer A  │ │Consumer B│ │Consumer N│
│(Alerting)  │ │(Storage) │ │(Analysis)│
└────────────┘ └──────────┘ └─────────┘
```

## Deployment Patterns

### 1. Standalone Mode
Single binary deployment for specific tasks:
```bash
# Prometheus metrics only
prysm local-producer ops-log --prometheus --prometheus-port 8080

# Disk health monitoring
prysm local-producer disk-health-metrics --prometheus
```

### 2. Distributed Mode
Multiple producers and consumers with NATS:
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Producer 1  ├────►│    NATS     ├────►│ Consumer 1  │
│ (Node A)    │     │  (Central)  │     │ (Monitoring)│
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
┌─────────────┐            │            ┌─────────────┐
│ Producer 2  ├────────────┤            │ Consumer 2  │
│ (Node B)    │            └───────────►│ (Alerting)  │
└─────────────┘                         └─────────────┘
```

### 3. Kubernetes Sidecar Pattern
Automatic injection via mutating webhook:
```
┌──────────────────────────────────────┐
│ Pod: rook-ceph-rgw                   │
│ ┌──────────────┐  ┌───────────────┐ │
│ │ RadosGW      │  │ Prysm Sidecar │ │
│ │ Container    │  │ (ops-log)     │ │
│ │              │  │               │ │
│ │ writes logs ─┼─►│ reads logs    │ │
│ └──────────────┘  │ exposes :9090 │ │
│                   └───────┬───────┘ │
└───────────────────────────┼─────────┘
                            │
                     ┌──────▼───────┐
                     │ Prometheus   │
                     │ (Scrapes)    │
                     └──────────────┘
```

## Technology Stack

### Core Technologies
- **Language**: Go 1.26
- **Messaging**: NATS 2.12+
- **Metrics**: Prometheus client_golang
- **Logging**: zerolog
- **CLI Framework**: Cobra + Viper
- **Configuration**: YAML with environment variable overrides

### Integration Technologies
- **Ceph Integration**: go-ceph library
- **AWS SDK**: S3 API compatibility
- **RabbitMQ**: amqp091-go (audit trail)
- **System Monitoring**: gopsutil, shirou/gopsutil

### Storage Layer (RadosGW Usage)
- **State Management**: NATS KV (Key-Value) store
- **Reconciliation**: Periodic sync with RadosGW state
- **Data Sources**: RadosGW Admin API

## Scalability Considerations

### Horizontal Scaling
- Multiple producer instances per node type
- Multiple consumer instances for load distribution
- NATS clustering for high availability

### Vertical Scaling
- Adjustable metric granularity
- Configurable collection intervals
- Optional metric types (enable only what's needed)

### Performance Optimizations
- **Dedicated Storage Maps**: Each metric type has optimized storage
- **Zero-Copy Processing**: Direct metric updates without aggregation
- **Efficient Memory Usage**: Only enabled metrics consume memory
- **Buffered Channels**: Non-blocking publishing
- **Fire-and-Forget**: Audit events don't block processing

## Security Architecture

### Authentication & Authorization
- NATS: Token-based authentication support
- RadosGW API: S3 access key + secret key
- RabbitMQ: AMQP authentication
- Kubernetes: RBAC for webhook and sidecar injection

### Data Protection
- TLS support for NATS connections
- Webhook TLS via cert-manager
- Sensitive credentials via Kubernetes Secrets
- Audit trail for compliance (CADF format)

### Multi-Tenancy
- Tenant isolation in metrics
- Bucket name collision prevention
- Per-tenant aggregation levels
- Domain/project scoping

## Monitoring & Observability

### Self-Monitoring
Prysm itself exposes metrics about its operation:
- Message processing rates
- Error counts
- Connection status
- Processing latency
- Queue depths

### Health Checks
- Prometheus `/metrics` endpoint
- NATS connection health
- File watcher status
- Disk accessibility

## Extension Points

### Custom Producers
Implement the producer interface:
```go
type Producer interface {
    Start(ctx context.Context) error
    Stop() error
    PublishToNATS(subject string, data []byte) error
    ExposePrometheusMetrics(port int) error
}
```

### Custom Consumers
Subscribe to NATS subjects and implement custom logic:
```go
type Consumer interface {
    Subscribe(subject string) error
    Process(data []byte) error
    PublishAlerts() error
}
```

## Best Practices

### Producer Configuration
- Enable only necessary metrics
- Use appropriate aggregation levels
- Configure reasonable collection intervals
- Monitor memory usage

### Consumer Configuration
- Use durable subscribers for critical data
- Implement proper error handling
- Configure appropriate timeouts
- Monitor subscription lag

### NATS Configuration
- Use JetStream for persistence
- Configure appropriate retention policies
- Monitor queue depths
- Set up clustering for HA

## Future Architecture Evolution

### Planned Enhancements
- Additional producer types (network metrics, etc.)
- Enhanced consumer capabilities
- Machine learning integration for anomaly detection
- Advanced correlation capabilities
- Enhanced visualization tools

### Extensibility Goals
- Plugin architecture for custom producers/consumers
- GraphQL API for data queries
- Unified configuration management
- Enhanced multi-cluster support
