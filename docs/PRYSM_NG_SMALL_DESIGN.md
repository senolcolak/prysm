# Prysm-NG-Small: Minimal Footprint Design Document

**Version:** 1.0
**Status:** Design Proposal
**Date:** March 5, 2026
**Philosophy:** "Do one thing extremely well, configure everything"

---

## Executive Summary

**Prysm-NG-Small** is a radical minimalist redesign inspired by Vector's philosophy: ultra-lightweight, single-binary, configuration-driven observability agent specifically for Ceph/RadosGW environments. The entire footprint is <15MB with <50MB RAM usage, making it suitable for sidecar deployment at scale.

**Core Philosophy:**
- **Minimal Footprint**: Single 15MB binary, 50MB RAM
- **Configuration-Driven**: 100% behavior defined via YAML
- **Zero Dependencies**: No external services required (optional NATS/Prometheus)
- **Fast**: <1ms processing latency per event
- **Fail-Safe**: Graceful degradation, never crash
- **Vector-Like**: Simple, predictable, observable

**Comparison:**

| Aspect | Prysm v1 | Prysm-NG (Full) | **Prysm-NG-Small** |
|--------|----------|-----------------|-------------------|
| Binary Size | ~20MB | ~40MB | **<15MB** |
| Memory | 256-512MB | 512MB-2GB | **<50MB** |
| Dependencies | NATS optional | NATS, PostgreSQL, etcd | **None (all optional)** |
| Complexity | Medium | High | **Minimal** |
| Config Lines | ~100 | ~500 | **~50** |
| Startup Time | ~5s | ~10s | **<1s** |
| Use Case | Testing | Enterprise | **Edge/Scale** |

**Timeline:** 6-9 months to GA (vs. 12-15 for full NG)

---

## Table of Contents

1. [Design Philosophy](#1-design-philosophy)
2. [Architecture](#2-architecture)
3. [Configuration System](#3-configuration-system)
4. [Core Pipeline](#4-core-pipeline)
5. [Minimal Components](#5-minimal-components)
6. [Memory Management](#6-memory-management)
7. [Performance](#7-performance)
8. [Deployment](#8-deployment)
9. [Comparison with Vector](#9-comparison-with-vector)
10. [Implementation Roadmap](#10-implementation-roadmap)

---

## 1. Design Philosophy

### 1.1 Inspired by Vector

**What We Learn from Vector:**
- Single binary, zero runtime dependencies
- Configuration as code (YAML/TOML)
- Pipeline model: Sources → Transforms → Sinks
- Predictable resource usage
- Observable by default
- Fast compilation to Rust (we'll use Go with optimizations)

**Our Specialization:**
- Purpose-built for Ceph/RadosGW
- S3 operation log expertise
- SMART data normalization
- Ceph-specific enrichment

### 1.2 Core Principles

#### Principle 1: Minimal by Default
```
Everything is optional except:
  1. Configuration file
  2. One source
  3. One sink

No databases, no state stores, no coordination services.
If you need them, you configure them.
```

#### Principle 2: Configuration > Code
```
All behavior is configuration.
Adding a feature = adding a config option, not writing code.
Complex processing = config pipelines, not custom code.
```

#### Principle 3: Predictable Resources
```
Memory: Capped and configurable (default: 50MB)
CPU: Single core sufficient (multi-core optional)
Disk: Only for buffering (optional)
Network: Minimal, batched
```

#### Principle 4: Fail-Safe Always
```
No panic(), no log.Fatal(), no os.Exit()
Errors = logs + metrics + optional degradation
Default behavior: drop data rather than crash
```

#### Principle 5: Observable
```
Self-monitoring is not optional.
Every component exposes metrics.
Every error is logged.
Every drop is counted.
```

---

## 2. Architecture

### 2.1 Pipeline Model (Vector-Inspired)

```
┌────────────────────────────────────────────────────────────┐
│                    Configuration                            │
│                     (YAML only)                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                  Pipeline Engine                            │
│                                                             │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐        │
│  │ Sources  │─────▶│Transforms│─────▶│  Sinks   │        │
│  │          │      │          │      │          │        │
│  │ • File   │      │ • Filter │      │ • NATS   │        │
│  │ • Exec   │      │ • Parse  │      │ • Prom   │        │
│  │ • HTTP   │      │ • Enrich │      │ • File   │        │
│  └──────────┘      │ • Sample │      │ • HTTP   │        │
│                    └──────────┘      └──────────┘        │
│                                                             │
│  ┌────────────────────────────────────────────┐           │
│  │         Ring Buffer (Memory Only)          │           │
│  │            Size: Configurable              │           │
│  └────────────────────────────────────────────┘           │
└────────────────────────────────────────────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │  Internal Metrics   │
              │  (Prometheus /metrics)│
              └─────────────────────┘
```

**Key Characteristics:**
- **No persistence**: All in-memory (optional disk buffer)
- **No coordination**: Single instance, no clustering
- **No external dependencies**: Self-contained
- **Stateless**: Restart = fresh start (by design)

### 2.2 Component Size Budget

| Component | Memory | Binary Size | Justification |
|-----------|--------|-------------|---------------|
| Core Engine | 10MB | 4MB | Pipeline + routing |
| Sources | 5MB | 2MB | File watcher, exec, HTTP |
| Transforms | 15MB | 4MB | Parsing, enrichment |
| Sinks | 10MB | 3MB | NATS, Prometheus, outputs |
| Config Parser | 5MB | 1MB | YAML parsing |
| Metrics | 5MB | 1MB | Self-monitoring |
| **Total** | **50MB** | **15MB** | **Target** |

---

## 3. Configuration System

### 3.1 Vector-Style Configuration

**Single YAML file. No API, no etcd, no database.**

```yaml
# prysm-ng-small.yaml
# Minimal configuration for ops-log monitoring

# Global settings (optional)
global:
  # Resource limits
  memory_limit: 50MB      # Hard limit, crash if exceeded
  cpu_limit: 1.0          # Max CPU cores

  # Logging
  log_level: info         # debug, info, warn, error
  log_format: json        # json or text

# Data directory (optional, for disk buffering)
data_dir: /var/lib/prysm-ng-small

# Sources: Where data comes from
sources:
  # Watch Ceph RGW operations log
  ops_log:
    type: file
    path: /var/log/ceph/ops-log.log

    # File watching
    read_from: end         # start, end, or beginning
    max_line_bytes: 102400 # 100KB max line size

    # Decoding
    decoding:
      codec: json

    # Fingerprinting (track file position)
    fingerprint:
      strategy: device_and_inode

# Transforms: Process data
transforms:
  # Parse and filter
  parse_ops_log:
    type: remap           # Vector-compatible transform
    inputs: [ops_log]
    source: |
      # Parse timestamp
      .timestamp = parse_timestamp!(.time, format: "%+")

      # Filter anonymous users (optional)
      if .user == "anonymous" {
        abort
      }

      # Extract tenant from user (format: tenant$user)
      .tenant = split(.user, "$")[0] ?? "default"
      .user_name = split(.user, "$")[1] ?? .user

      # Add instance metadata
      .instance_id = "${HOSTNAME}"
      .cluster = "${CLUSTER_ID}"

      # Convert latency from ms to seconds
      .latency_seconds = to_float(.total_time) / 1000.0

      # Categorize operation
      .operation_type = if includes(["PUT", "POST"], .method) {
        "write"
      } else if includes(["GET", "HEAD"], .method) {
        "read"
      } else if includes(["DELETE"], .method) {
        "delete"
      } else {
        "other"
      }

      # Categorize errors
      if .http_status >= 400 {
        .error_category = if .http_status >= 500 {
          "server_error"
        } else if includes([408, 504, 598, 499], .http_status) {
          "timeout_error"
        } else {
          "client_error"
        }
      }

  # Sample for high-volume environments (optional)
  sample_logs:
    type: sample
    inputs: [parse_ops_log]
    rate: 10              # Keep 1 in 10 events (configurable)
    exclude:              # Never sample these
      field: http_status
      values: [500, 502, 503, 504]  # Always keep errors

# Sinks: Where data goes
sinks:
  # Prometheus metrics
  prometheus:
    type: prometheus_exporter
    inputs: [parse_ops_log]
    address: 0.0.0.0:9090

    # Metrics to generate (configuration-driven)
    metrics:
      # Request counter
      - type: counter
        name: radosgw_requests_total
        labels:
          tenant: "{{ tenant }}"
          bucket: "{{ bucket }}"
          method: "{{ method }}"
          status: "{{ http_status }}"

      # Latency histogram
      - type: histogram
        name: radosgw_request_duration_seconds
        field: latency_seconds
        labels:
          tenant: "{{ tenant }}"
          method: "{{ method }}"
        buckets: [0.001, 0.01, 0.1, 0.5, 1.0, 5.0]

      # Bytes transferred
      - type: counter
        name: radosgw_bytes_sent_total
        field: bytes_sent
        labels:
          tenant: "{{ tenant }}"
          bucket: "{{ bucket }}"

      - type: counter
        name: radosgw_bytes_received_total
        field: bytes_received
        labels:
          tenant: "{{ tenant }}"
          bucket: "{{ bucket }}"

      # Errors by category
      - type: counter
        name: radosgw_errors_total
        labels:
          tenant: "{{ tenant }}"
          category: "{{ error_category }}"
        condition: "http_status >= 400"

  # NATS (optional)
  nats_events:
    type: nats
    inputs: [parse_ops_log]
    url: nats://nats:4222
    subject: ops.log.{{ tenant }}

    # Encoding
    encoding:
      codec: json

    # Batching for efficiency
    batch:
      max_events: 100
      timeout_secs: 1

    # Error handling
    healthcheck:
      enabled: true
    buffer:
      type: memory
      max_events: 10000
      when_full: drop_newest  # or block

  # Console output (debug)
  console:
    type: console
    inputs: [parse_ops_log]
    encoding:
      codec: json
      only_fields: [timestamp, user, bucket, method, http_status, latency_seconds]

    # Only enable in debug mode
    enabled: false
```

**Configuration size: ~100 lines (vs. 500+ in full NG)**

### 3.2 Minimal Configuration (Absolute Minimum)

```yaml
# Bare minimum: 15 lines
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

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [parse]
    address: 0.0.0.0:9090
```

**This is the entire config needed for basic metrics!**

### 3.3 Disk Health Monitoring Example

```yaml
# Disk health monitoring (30 lines)
sources:
  disk_health:
    type: exec
    command: ["/usr/sbin/smartctl", "-A", "-j", "/dev/sda"]
    streaming: false
    interval: 60  # Run every 60 seconds

transforms:
  parse_smart:
    type: remap
    inputs: [disk_health]
    source: |
      . = parse_json!(.message)

      # Extract SMART attributes
      .temperature = .temperature.current ?? 0
      .reallocated_sectors = .ata_smart_attributes.table[4].raw.value ?? 0
      .power_on_hours = .power_on_time.hours ?? 0

      # Add metadata
      .disk = "/dev/sda"
      .node = "${HOSTNAME}"

sinks:
  disk_metrics:
    type: prometheus_exporter
    inputs: [parse_smart]
    address: 0.0.0.0:9091

    metrics:
      - type: gauge
        name: disk_temperature_celsius
        field: temperature
        labels:
          disk: "{{ disk }}"
          node: "{{ node }}"

      - type: gauge
        name: disk_reallocated_sectors
        field: reallocated_sectors
        labels:
          disk: "{{ disk }}"
```

### 3.4 Configuration Validation

```bash
# Validate configuration before deploying
prysm-ng-small validate prysm-ng-small.yaml

# Output:
✓ Configuration is valid
✓ All sources defined
✓ All transforms have valid inputs
✓ All sinks have valid inputs
✓ Estimated memory usage: 45MB
✓ No circular dependencies

# Test configuration with sample data
prysm-ng-small test prysm-ng-small.yaml \
  --input sample-log.json

# Hot-reload configuration (USR1 signal)
kill -USR1 $(pidof prysm-ng-small)
```

---

## 4. Core Pipeline

### 4.1 Pipeline Engine

```go
// Minimal pipeline engine (~200 lines total)
type Pipeline struct {
    sources    map[string]Source
    transforms map[string]Transform
    sinks      map[string]Sink

    // Ring buffer for events
    buffer     *RingBuffer

    // Metrics
    metrics    *Metrics

    // Config
    config     *Config
}

type Event struct {
    Data      map[string]interface{}
    Metadata  map[string]string
    Timestamp time.Time
}

func (p *Pipeline) Run(ctx context.Context) error {
    // Start sources
    for name, source := range p.sources {
        go p.runSource(ctx, name, source)
    }

    // Process events
    go p.processEvents(ctx)

    <-ctx.Done()
    return nil
}

func (p *Pipeline) runSource(ctx context.Context, name string, source Source) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            events, err := source.Read(ctx)
            if err != nil {
                p.metrics.RecordSourceError(name, err)
                time.Sleep(1 * time.Second)
                continue
            }

            for _, event := range events {
                // Add to buffer
                if !p.buffer.Add(event) {
                    p.metrics.RecordDropped(name, "buffer_full")
                }
            }
        }
    }
}

func (p *Pipeline) processEvents(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Read from buffer
            event, ok := p.buffer.Get()
            if !ok {
                time.Sleep(10 * time.Millisecond)
                continue
            }

            // Apply transforms
            for _, transform := range p.transforms {
                event = transform.Process(event)
                if event == nil {
                    break  // Event filtered out
                }
            }

            if event == nil {
                continue
            }

            // Send to sinks
            for name, sink := range p.sinks {
                if err := sink.Write(event); err != nil {
                    p.metrics.RecordSinkError(name, err)
                }
            }
        }
    }
}
```

### 4.2 Source Interface (Minimal)

```go
type Source interface {
    Read(ctx context.Context) ([]Event, error)
}

// File source (~100 lines)
type FileSource struct {
    path     string
    watcher  *fsnotify.Watcher
    reader   *bufio.Reader
    decoder  Decoder
    position int64
}

// Exec source (~50 lines)
type ExecSource struct {
    command  []string
    interval time.Duration
    decoder  Decoder
}

// HTTP source (~80 lines)
type HTTPSource struct {
    address string
    path    string
    decoder Decoder
}
```

### 4.3 Transform Interface (Minimal)

```go
type Transform interface {
    Process(event Event) Event  // nil = filtered out
}

// Remap transform (Vector-compatible VRL subset)
type RemapTransform struct {
    script *Script
}

// Filter transform
type FilterTransform struct {
    condition Condition
}

// Sample transform
type SampleTransform struct {
    rate     int
    exclude  Condition
}
```

### 4.4 Sink Interface (Minimal)

```go
type Sink interface {
    Write(event Event) error
}

// Prometheus sink (~150 lines)
type PrometheusSink struct {
    address  string
    registry *prometheus.Registry
    metrics  map[string]prometheus.Collector
}

// NATS sink (~100 lines)
type NATSSink struct {
    conn    *nats.Conn
    subject string
    encoder Encoder
    batch   *Batcher
}

// Console sink (~30 lines)
type ConsoleSink struct {
    writer  io.Writer
    encoder Encoder
}
```

---

## 5. Minimal Components

### 5.1 Only Essential Components

**Included:**
- ✅ File source (ops-log)
- ✅ Exec source (smartctl)
- ✅ HTTP source (webhook)
- ✅ Remap transform (VRL subset)
- ✅ Filter transform
- ✅ Sample transform
- ✅ Prometheus sink
- ✅ NATS sink
- ✅ Console sink

**Explicitly NOT Included:**
- ❌ No leader election
- ❌ No state storage
- ❌ No coordination
- ❌ No HA (deploy multiple instances if needed)
- ❌ No persistence (restart = fresh start)
- ❌ No plugin system (keep it simple)
- ❌ No GUI/API (config file only)
- ❌ No complex stream processing (use transforms)

### 5.2 Error Handling (Configuration-Driven)

```yaml
# Global error handling
global:
  error_handling:
    # What to do on errors
    on_error: log_and_continue  # or drop, or retry

    # Retry configuration
    retry:
      max_attempts: 3
      initial_delay: 100ms
      max_delay: 5s

    # Health checks
    healthcheck:
      enabled: true
      interval: 30s
```

**Implementation:**
```go
func (p *Pipeline) handleError(err error, component string) {
    switch p.config.ErrorHandling.OnError {
    case "log_and_continue":
        log.Error().Err(err).Str("component", component).Msg("Error occurred")
        p.metrics.RecordError(component)

    case "drop":
        p.metrics.RecordDropped(component, "error")

    case "retry":
        // Retry with backoff
        p.retryWithBackoff(component)
    }

    // Never panic or exit
}
```

---

## 6. Memory Management

### 6.1 Memory Budget

```go
type MemoryManager struct {
    limit      int64  // Hard limit (e.g., 50MB)
    current    int64  // Current usage
    ringBuffer *RingBuffer
}

func (m *MemoryManager) Allocate(size int64) error {
    if m.current + size > m.limit {
        // Apply back-pressure
        return errors.New("memory limit reached")
    }
    m.current += size
    return nil
}

// Ring buffer with fixed size
type RingBuffer struct {
    events   []Event
    capacity int
    head     int
    tail     int
    size     int
}

func (r *RingBuffer) Add(event Event) bool {
    if r.size >= r.capacity {
        // Buffer full, drop oldest or newest based on config
        if r.config.WhenFull == "drop_oldest" {
            r.tail = (r.tail + 1) % r.capacity
        } else {
            return false  // drop_newest
        }
    }

    r.events[r.head] = event
    r.head = (r.head + 1) % r.capacity
    r.size++
    return true
}
```

### 6.2 Configuration

```yaml
global:
  # Memory configuration
  memory:
    limit: 50MB              # Hard limit
    buffer_size: 10000       # Events in ring buffer

    # When buffer is full
    when_full: drop_oldest   # or drop_newest, or block
```

---

## 7. Performance

### 7.1 Performance Targets

| Metric | Target | Current Prysm v1 |
|--------|--------|------------------|
| **Binary Size** | <15MB | ~20MB |
| **Memory (Idle)** | <20MB | ~100MB |
| **Memory (Active)** | <50MB | ~256MB |
| **Startup Time** | <1s | ~5s |
| **Processing Latency** | <1ms/event | ~5-10ms |
| **Throughput** | 100K events/s | ~10K events/s |
| **CPU (Idle)** | <1% | ~5% |
| **CPU (Active)** | <100% (1 core) | ~200% |

### 7.2 Optimization Techniques

**1. Zero-Copy Where Possible**
```go
// Avoid copying event data
type Event struct {
    Data map[string]interface{}  // Reference, not copy
}
```

**2. Object Pooling**
```go
var eventPool = sync.Pool{
    New: func() interface{} {
        return &Event{
            Data: make(map[string]interface{}, 16),
        }
    },
}

func getEvent() *Event {
    return eventPool.Get().(*Event)
}

func putEvent(e *Event) {
    // Clear and return to pool
    for k := range e.Data {
        delete(e.Data, k)
    }
    eventPool.Put(e)
}
```

**3. Batching**
```go
type Batcher struct {
    maxEvents int
    timeout   time.Duration
    batch     []Event
}

func (b *Batcher) Add(event Event) {
    b.batch = append(b.batch, event)
    if len(b.batch) >= b.maxEvents {
        b.Flush()
    }
}
```

**4. Minimal Allocations**
```go
// Pre-allocate buffers
type FileSource struct {
    buffer []byte  // Reuse for each read
}

func (s *FileSource) Read(ctx context.Context) ([]Event, error) {
    // Reuse buffer instead of allocating
    n, err := s.file.Read(s.buffer)
    // ...
}
```

### 7.3 Benchmarks (Target)

```
BenchmarkPipeline/file_to_prometheus-8     100000    1000 ns/op    0 allocs/op
BenchmarkPipeline/file_to_nats-8           80000     1200 ns/op    0 allocs/op
BenchmarkTransform/remap-8                 500000    500 ns/op     0 allocs/op
BenchmarkSink/prometheus-8                 200000    800 ns/op     1 allocs/op
```

---

## 8. Deployment

### 8.1 Kubernetes Sidecar (Primary Use Case)

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: rgw-with-prysm
spec:
  containers:
    # Main container
    - name: radosgw
      image: ceph/daemon:latest
      volumeMounts:
        - name: logs
          mountPath: /var/log/ceph

    # Prysm sidecar (minimal footprint)
    - name: prysm
      image: ghcr.io/prysm/prysm-ng-small:latest

      # Resource limits (small!)
      resources:
        requests:
          cpu: 50m        # 50 millicores
          memory: 32Mi    # 32 MB
        limits:
          cpu: 200m       # 200 millicores max
          memory: 64Mi    # 64 MB max

      # Configuration
      volumeMounts:
        - name: config
          mountPath: /etc/prysm
        - name: logs
          mountPath: /var/log/ceph
          readOnly: true

      # Command
      command:
        - /prysm-ng-small
        - --config
        - /etc/prysm/config.yaml

      # Metrics port
      ports:
        - containerPort: 9090
          name: metrics

  volumes:
    - name: logs
      emptyDir: {}

    - name: config
      configMap:
        name: prysm-config
```

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prysm-config
data:
  config.yaml: |
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

    sinks:
      metrics:
        type: prometheus_exporter
        inputs: [parse]
        address: 0.0.0.0:9090
```

**That's it! 32MB RAM, 50m CPU per sidecar.**

### 8.2 Standalone Deployment

```bash
# Download binary (15MB)
curl -L https://github.com/prysm/releases/download/v2.0.0/prysm-ng-small-linux-amd64 \
  -o /usr/local/bin/prysm-ng-small

chmod +x /usr/local/bin/prysm-ng-small

# Create config
cat > /etc/prysm/config.yaml << 'EOF'
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

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [parse]
    address: 0.0.0.0:9090
EOF

# Run
prysm-ng-small --config /etc/prysm/config.yaml
```

### 8.3 Docker

```dockerfile
FROM scratch

# Copy single binary
COPY prysm-ng-small /

# Expose metrics port
EXPOSE 9090

# Run
ENTRYPOINT ["/prysm-ng-small"]
CMD ["--config", "/etc/prysm/config.yaml"]
```

**Image size: ~16MB (binary + scratch base)**

```bash
docker run -d \
  --name prysm \
  -v /var/log/ceph:/var/log/ceph:ro \
  -v $(pwd)/config.yaml:/etc/prysm/config.yaml:ro \
  -p 9090:9090 \
  --memory=64m \
  --cpus=0.2 \
  prysm-ng-small:latest
```

### 8.4 DaemonSet for Node Monitoring

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: prysm-node-monitor
spec:
  selector:
    matchLabels:
      app: prysm-node-monitor
  template:
    metadata:
      labels:
        app: prysm-node-monitor
    spec:
      containers:
      - name: prysm
        image: ghcr.io/prysm/prysm-ng-small:latest

        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 200m
            memory: 64Mi

        volumeMounts:
        - name: config
          mountPath: /etc/prysm
        - name: dev
          mountPath: /dev
          readOnly: true

        securityContext:
          privileged: true  # For smartctl access

      volumes:
      - name: config
        configMap:
          name: prysm-disk-config
      - name: dev
        hostPath:
          path: /dev
```

---

## 9. Comparison with Vector

### 9.1 What We Adopt from Vector

| Feature | Vector | Prysm-NG-Small |
|---------|--------|----------------|
| **Single Binary** | ✅ Yes | ✅ Yes |
| **Config-Driven** | ✅ YAML/TOML | ✅ YAML only |
| **Pipeline Model** | ✅ Yes | ✅ Yes |
| **VRL (Remap Language)** | ✅ Full | ✅ Subset |
| **Minimal Footprint** | ✅ ~30MB | ✅ <15MB |
| **Memory Limit** | ✅ Configurable | ✅ Configurable |
| **No Dependencies** | ✅ Yes | ✅ Yes |
| **Fast** | ✅ Rust | ✅ Go (optimized) |

### 9.2 What We Add (Ceph-Specific)

| Feature | Vector | Prysm-NG-Small |
|---------|--------|----------------|
| **S3 Log Parsing** | ⚠️ Generic | ✅ Built-in |
| **SMART Normalization** | ❌ No | ✅ Yes |
| **Ceph OSD Mapping** | ❌ No | ✅ Yes |
| **Tenant Extraction** | ⚠️ Manual | ✅ Automatic |
| **CADF Audit Format** | ❌ No | ✅ Optional |
| **Ceph Enrichment** | ❌ No | ✅ Yes |

### 9.3 What We Don't Need (Keeping It Minimal)

| Feature | Vector | Prysm-NG-Small |
|---------|--------|----------------|
| **Complex Routing** | ✅ Yes | ❌ Simple only |
| **Many Sources** | ✅ 20+ | ⚠️ 3 (file, exec, http) |
| **Many Sinks** | ✅ 30+ | ⚠️ 3 (prom, nats, console) |
| **Lua Scripting** | ✅ Yes | ❌ No |
| **Clustering** | ⚠️ Limited | ❌ No |
| **Persistence** | ✅ Disk buffer | ⚠️ Optional |

### 9.4 Vector Configuration Comparison

**Vector (Generic Log Processing):**
```toml
[sources.logs]
type = "file"
include = ["/var/log/app.log"]

[transforms.parse]
type = "remap"
inputs = ["logs"]
source = '''
  . = parse_json!(.message)
'''

[sinks.prometheus]
type = "prometheus_exporter"
inputs = ["parse"]
address = "0.0.0.0:9090"
```

**Prysm-NG-Small (Same Thing, Ceph-Optimized):**
```yaml
sources:
  logs:
    type: file
    path: /var/log/ceph/ops-log.log
    decoding:
      codec: json  # Built-in JSON parsing

transforms:
  enrich:
    type: remap
    inputs: [logs]
    source: |
      # Ceph-specific enrichment
      .tenant = split(.user, "$")[0] ?? "default"

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [enrich]
    address: 0.0.0.0:9090
```

**Both achieve the same goal. Prysm adds Ceph knowledge.**

---

## 10. Implementation Roadmap

### Phase 1: Core Engine (Months 1-2)

**Goal:** Minimal working pipeline

**Deliverables:**
- ✅ Configuration parser (YAML)
- ✅ Pipeline engine (sources → transforms → sinks)
- ✅ Ring buffer with memory limits
- ✅ File source
- ✅ Remap transform (VRL subset)
- ✅ Prometheus sink
- ✅ Console sink

**Binary Size Target:** 10MB
**Memory Target:** 30MB
**Test Coverage:** 60%

### Phase 2: Essential Components (Months 3-4)

**Goal:** Production-ready for ops-log

**Deliverables:**
- ✅ Exec source (smartctl)
- ✅ HTTP source (webhooks)
- ✅ NATS sink
- ✅ Filter transform
- ✅ Sample transform
- ✅ Error handling framework
- ✅ Self-monitoring metrics

**Binary Size Target:** 12MB
**Memory Target:** 40MB
**Test Coverage:** 70%

### Phase 3: Ceph Specialization (Months 5-6)

**Goal:** Ceph-specific features

**Deliverables:**
- ✅ S3 log parsing helpers
- ✅ Tenant extraction logic
- ✅ SMART attribute normalization
- ✅ Ceph OSD mapping
- ✅ CADF audit format (optional)
- ✅ Enrichment transforms

**Binary Size Target:** 14MB
**Memory Target:** 45MB
**Test Coverage:** 80%

### Phase 4: Polish & Optimize (Months 7-8)

**Goal:** Performance and stability

**Deliverables:**
- ✅ Performance optimizations
- ✅ Memory profiling and reduction
- ✅ Binary size optimization
- ✅ Comprehensive testing
- ✅ Documentation
- ✅ Examples

**Binary Size Target:** <15MB
**Memory Target:** <50MB
**Test Coverage:** 85%

### Phase 5: GA Release (Month 9)

**Goal:** Production release

**Deliverables:**
- ✅ Security audit
- ✅ Load testing
- ✅ Deployment guides
- ✅ Migration from Prysm v1
- ✅ GA release

**Final Targets:**
- Binary: <15MB ✅
- Memory: <50MB ✅
- Startup: <1s ✅
- Throughput: 100K events/s ✅

---

## 11. Configuration Examples

### 11.1 Ops-Log Monitoring (Production)

```yaml
# Complete ops-log monitoring configuration
sources:
  ops_log:
    type: file
    path: /var/log/ceph/ops-log.log
    read_from: end
    decoding:
      codec: json
    fingerprint:
      strategy: device_and_inode

transforms:
  parse_and_enrich:
    type: remap
    inputs: [ops_log]
    source: |
      # Parse timestamp
      .timestamp = parse_timestamp!(.time, format: "%+")

      # Extract tenant and user
      parts = split(.user, "$")
      .tenant = parts[0] ?? "default"
      .user_name = parts[1] ?? .user

      # Convert latency
      .latency_seconds = to_float(.total_time) / 1000.0

      # Add metadata
      .cluster = "${CLUSTER_ID}"
      .instance = "${HOSTNAME}"

      # Categorize
      .operation_type = if includes(["PUT", "POST"], .method) {
        "write"
      } else if includes(["GET", "HEAD"], .method) {
        "read"
      } else {
        "other"
      }

  filter_anonymous:
    type: filter
    inputs: [parse_and_enrich]
    condition: .user != "anonymous"

  sample_reads:
    type: sample
    inputs: [filter_anonymous]
    rate: 10
    exclude:
      field: operation_type
      values: [write, delete]  # Don't sample writes/deletes

sinks:
  prometheus:
    type: prometheus_exporter
    inputs: [sample_reads]
    address: 0.0.0.0:9090
    metrics:
      - type: counter
        name: radosgw_requests_total
        labels:
          tenant: "{{ tenant }}"
          bucket: "{{ bucket }}"
          method: "{{ method }}"
          status: "{{ http_status }}"

      - type: histogram
        name: radosgw_request_duration_seconds
        field: latency_seconds
        labels:
          method: "{{ method }}"
        buckets: [0.001, 0.01, 0.1, 0.5, 1.0, 5.0]

      - type: counter
        name: radosgw_bytes_total
        field: bytes_sent
        labels:
          tenant: "{{ tenant }}"
          direction: sent

      - type: counter
        name: radosgw_bytes_total
        field: bytes_received
        labels:
          tenant: "{{ tenant }}"
          direction: received

  nats:
    type: nats
    inputs: [sample_reads]
    url: nats://nats:4222
    subject: ops.log.{{ tenant }}
    batch:
      max_events: 100
      timeout_secs: 1
    buffer:
      type: memory
      max_events: 10000
      when_full: drop_oldest
```

### 11.2 Disk Health Monitoring

```yaml
sources:
  # Monitor multiple disks
  disk_sda:
    type: exec
    command: ["/usr/sbin/smartctl", "-A", "-j", "/dev/sda"]
    interval: 60

  disk_sdb:
    type: exec
    command: ["/usr/sbin/smartctl", "-A", "-j", "/dev/sdb"]
    interval: 60

  disk_nvme0:
    type: exec
    command: ["/usr/sbin/nvme", "smart-log", "-o", "json", "/dev/nvme0n1"]
    interval: 60

transforms:
  parse_smart:
    type: remap
    inputs: [disk_sda, disk_sdb]
    source: |
      . = parse_json!(.message)

      # Normalize SMART attributes
      .temperature = .temperature.current ?? 0
      .reallocated = .ata_smart_attributes.table[4].raw.value ?? 0
      .power_on_hours = .power_on_time.hours ?? 0

      # Extract disk from command
      .disk = split(.command, " ")[2]
      .node = "${HOSTNAME}"

  parse_nvme:
    type: remap
    inputs: [disk_nvme0]
    source: |
      . = parse_json!(.message)

      # NVMe attributes
      .temperature = .temperature ?? 0
      .available_spare = .available_spare ?? 100
      .critical_warning = .critical_warning ?? 0

      .disk = "/dev/nvme0n1"
      .node = "${HOSTNAME}"

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [parse_smart, parse_nvme]
    address: 0.0.0.0:9091
    metrics:
      - type: gauge
        name: disk_temperature_celsius
        field: temperature
        labels:
          disk: "{{ disk }}"
          node: "{{ node }}"

      - type: gauge
        name: disk_reallocated_sectors
        field: reallocated
        labels:
          disk: "{{ disk }}"

      - type: gauge
        name: disk_power_on_hours
        field: power_on_hours
        labels:
          disk: "{{ disk }}"
```

### 11.3 Multi-Tenant with Per-Tenant Sampling

```yaml
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

  # Sample differently per tenant
  sample_tenant_a:
    type: sample
    inputs: [parse]
    rate: 1  # Keep all
    condition: .tenant == "tenant-a"  # High-value tenant

  sample_tenant_b:
    type: sample
    inputs: [parse]
    rate: 10  # Keep 1 in 10
    condition: .tenant == "tenant-b"  # Standard tenant

  sample_others:
    type: sample
    inputs: [parse]
    rate: 100  # Keep 1 in 100
    condition: .tenant != "tenant-a" && .tenant != "tenant-b"

sinks:
  metrics:
    type: prometheus_exporter
    inputs: [sample_tenant_a, sample_tenant_b, sample_others]
    address: 0.0.0.0:9090
```

---

## 12. Migration from Prysm v1

### 12.1 Feature Comparison

| Feature | Prysm v1 | Prysm-NG-Small |
|---------|----------|----------------|
| **Ops-Log Monitoring** | ✅ Yes | ✅ Yes |
| **Disk Health** | ✅ Yes | ✅ Yes |
| **Quota Monitoring** | ⚠️ Partial | ✅ Via config |
| **CADF Audit** | ✅ Yes | ✅ Optional |
| **NATS** | ✅ Yes | ✅ Yes |
| **Prometheus** | ✅ Yes | ✅ Yes |
| **Binary Size** | 20MB | <15MB ✅ |
| **Memory** | 256MB | <50MB ✅ |
| **Config** | ~100 lines | ~50 lines ✅ |
| **Complexity** | Medium | Low ✅ |

### 12.2 Migration Steps

**Step 1: Install Prysm-NG-Small**
```bash
# Download
curl -L https://github.com/prysm/releases/download/v2.0.0/prysm-ng-small \
  -o /usr/local/bin/prysm-ng-small

chmod +x /usr/local/bin/prysm-ng-small
```

**Step 2: Convert Configuration**
```bash
# Auto-convert v1 config to NG-Small
prysm-ng-small convert \
  --from prysm-v1-config.yaml \
  --to prysm-ng-small-config.yaml
```

**Step 3: Run in Parallel**
```bash
# Run both for comparison
prysm-v1 --config old-config.yaml &
prysm-ng-small --config new-config.yaml &

# Compare metrics
diff <(curl -s localhost:8080/metrics | sort) \
     <(curl -s localhost:9090/metrics | sort)
```

**Step 4: Switch Over**
```bash
# Stop v1
pkill prysm-v1

# Keep only NG-Small
# (Already running from step 3)
```

---

## 13. Comparison Matrix

### Prysm v1 vs. NG (Full) vs. NG-Small

| Aspect | v1 | NG (Full) | **NG-Small** |
|--------|----|-----------|--------------|
| **Philosophy** | Specialized | Enterprise | **Minimal** |
| **Binary Size** | 20MB | 40MB | **<15MB ✅** |
| **Memory (Idle)** | 100MB | 200MB | **<20MB ✅** |
| **Memory (Active)** | 256MB | 512MB-2GB | **<50MB ✅** |
| **Startup Time** | 5s | 10s | **<1s ✅** |
| **Dependencies** | NATS opt. | Multiple | **None ✅** |
| **Configuration** | 100 lines | 500 lines | **50 lines ✅** |
| **Complexity** | Medium | High | **Low ✅** |
| **HA** | No | Yes | **No (deploy multiple)** |
| **Persistence** | No | Yes | **No (by design)** |
| **Plugins** | No | Yes | **No** |
| **State Storage** | No | PostgreSQL | **No** |
| **Coordination** | No | etcd | **No** |
| **Use Case** | Dev/Test | Enterprise | **Edge/Scale ✅** |
| **Timeline** | Done | 15 months | **9 months ✅** |
| **Ops Effort** | Medium | Low (automated) | **Lowest (config only) ✅** |

---

## 14. Success Criteria

### Technical Criteria

✅ **Footprint**
- Binary size < 15MB
- Memory usage < 50MB (active)
- Memory usage < 20MB (idle)
- Startup time < 1 second
- Single file deployment

✅ **Performance**
- Process 100K events/second
- <1ms latency per event
- <100% CPU (single core)
- Zero-copy where possible

✅ **Reliability**
- No log.Fatal() or panic()
- Graceful degradation always
- Predictable resource usage
- Clear error messages

✅ **Configuration**
- 100% behavior via YAML
- No external dependencies required
- Hot-reload support
- Validation before start

### Operational Criteria

✅ **Ease of Use**
- Single binary download
- <50 line minimal config
- No external services required
- Clear documentation

✅ **Observability**
- Self-monitoring built-in
- Prometheus /metrics endpoint
- Structured logging
- Clear status indicators

✅ **Compatibility**
- Works with existing Prysm v1 deployments
- Vector-compatible config style
- Standard Prometheus format
- NATS JetStream compatible

### Adoption Criteria

✅ **Target Users**
- Organizations with 100+ RGW pods (sidecar use case)
- Edge deployments with limited resources
- Cost-conscious deployments
- Simple monitoring needs

✅ **Success Metrics**
- <50MB RAM per instance
- Can run on 50m CPU
- Suitable for free-tier cloud instances
- Works on Raspberry Pi

---

## 15. FAQ

### Q: Why not just use Vector?

**A:** Vector is excellent for generic log processing, but:
- We need Ceph/RadosGW-specific knowledge built-in
- Our binary is smaller (<15MB vs. 30MB)
- We have Ceph-specific transforms (tenant extraction, SMART normalization)
- We're optimized for the exact use case (S3 ops logs, disk health)
- Configuration is simpler for Ceph users

### Q: What if I need HA?

**A:** Deploy multiple instances with a load balancer. Each instance is independent and stateless. No coordination needed.

### Q: What if I need persistence?

**A:** Use the NATS sink with JetStream. NATS provides persistence, not Prysm-NG-Small. This keeps Prysm minimal.

### Q: What if I need complex stream processing?

**A:** Use Prysm-NG (full version) or an external stream processor like Apache Flink. NG-Small is intentionally simple.

### Q: Can I run this on Raspberry Pi?

**A:** Yes! <15MB binary, <50MB RAM works fine on RPi 3+.

### Q: How is this different from Fluent Bit?

**A:** Similar philosophy (minimal footprint), but:
- We're specialized for Ceph/RadosGW
- Simpler configuration model
- Smaller binary (<15MB vs. ~20MB)
- No Lua scripting complexity
- Built-in Ceph knowledge

### Q: What about plugins?

**A:** No plugins. If you need extensibility, use Prysm-NG (full). NG-Small is intentionally simple.

### Q: Can I contribute?

**A:** Yes! We focus on:
- Performance improvements
- Memory optimizations
- Ceph-specific features
- Bug fixes
- Documentation

---

## 16. Conclusion

**Prysm-NG-Small** is a radical minimalist redesign focused on:
- **<15MB binary**: Deploy anywhere
- **<50MB RAM**: Run at scale (100s of instances)
- **Configuration-driven**: Ops team has full control
- **Zero dependencies**: Just a binary + config file
- **Fail-safe**: Never crash, always degrade
- **Fast**: <1ms per event, 100K events/sec

**Perfect For:**
- Kubernetes sidecar deployment (32MB RAM per pod)
- Edge deployments with limited resources
- Organizations with 100+ RGW instances
- Simple monitoring needs
- Cost-conscious deployments

**Not For:**
- Complex stream processing (use NG-Full)
- High Availability requirements (deploy multiple)
- Persistent state needs (use external storage)
- Plugin/extensibility needs (use NG-Full)

**Timeline:** 9 months to GA (vs. 15 for NG-Full)

**Next Steps:**
1. Review design with team
2. Prototype core pipeline (Month 1)
3. Benchmark against targets
4. Iterate based on feedback
5. Begin implementation

---

**Document Status:** DRAFT - Ready for Review
**Target Score:** 9/10 for edge/scale use cases
**Philosophy:** "Minimal footprint, maximum configuration"
