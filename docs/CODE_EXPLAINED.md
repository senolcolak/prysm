# How Prysm Works: Code Architecture and Flow

## Table of Contents
1. [Code Structure](#code-structure)
2. [Core Components](#core-components)
3. [Command-Line Interface](#command-line-interface)
4. [Producer Implementation](#producer-implementation)
5. [Consumer Implementation](#consumer-implementation)
6. [Data Flow and Processing](#data-flow-and-processing)
7. [Configuration Management](#configuration-management)
8. [Messaging Layer](#messaging-layer)
9. [Metrics Collection](#metrics-collection)
10. [Key Design Patterns](#key-design-patterns)

---

## Code Structure

### Directory Layout

```
prysm/
├── cmd/
│   └── main.go                    # Application entry point
├── pkg/
│   ├── commands/                  # CLI commands and subcommands
│   │   ├── ctl.go                # Root command setup
│   │   ├── consumers.go          # Consumer command group
│   │   ├── local_producer.go     # Local producer command group
│   │   ├── remote_producer.go    # Remote producer command group
│   │   └── producer_*.go         # Individual producer commands
│   ├── producers/                 # Producer implementations
│   │   ├── config/               # Shared producer configuration
│   │   ├── opslog/               # Operations log producer
│   │   ├── diskhealthmetrics/   # Disk health producer
│   │   ├── kernelmetrics/       # Kernel metrics producer
│   │   ├── resourceusage/       # Resource usage producer
│   │   ├── bucketnotify/        # Bucket notifications producer
│   │   ├── quotausagemonitor/   # Quota monitor producer
│   │   └── radosgwusage/        # RadosGW usage producer
│   └── consumer/                  # Consumer implementations
│       └── quotausageconsumer/   # Quota usage consumer
├── ops-log-k8s-mutating-wh/      # Kubernetes webhook for sidecar injection
│   ├── main.go                   # Webhook server entry point
│   └── webhook.go                # Mutation logic
├── examples/                      # Example configurations
├── docs/                          # Documentation
└── go.mod                        # Go module definition
```

---

## Core Components

### 1. Main Entry Point

**File**: `cmd/main.go`

```go
package main

import (
    "github.com/cobaltcore-dev/prysm/pkg/commands"
)

func main() {
    commands.Execute()
}
```

**How it works**:
- Ultra-simple entry point
- Delegates all logic to the commands package
- Allows for clean separation of concerns

### 2. Command Controller

**File**: `pkg/commands/ctl.go`

**Key Components**:

```go
var rootCmd = &cobra.Command{
    Use:   "prysm",
    Short: "CLI for Ceph & RadosGW observability",
    Long:  "A CLI tool to manage Ceph & RadosGW observability...",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return setUpLogs(v)
    },
}

func init() {
    // Check if running in Kubernetes pod
    runningInPod = checkIfRunningInPod()

    // Add global flags
    rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", "warn", "Log level")

    // Register subcommands
    rootCmd.AddCommand(consumerCmd)
    rootCmd.AddCommand(localProducerCmd)
    rootCmd.AddCommand(remoteProducerCmd)
}
```

**How it works**:
1. Uses Cobra for CLI framework
2. Sets up logging before any command execution
3. Detects Kubernetes environment automatically
4. Provides global flags (verbosity) to all subcommands
5. Organizes commands into logical groups (consumers, local producers, remote producers)

**Kubernetes Detection**:
```go
func checkIfRunningInPod() bool {
    // Check for Kubernetes service account files
    if _, err := os.Stat("/run/secrets/kubernetes.io/serviceaccount/ca.crt"); err == nil {
        if _, err := os.Stat("/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
            // Check for Kubernetes environment variables
            if _, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok {
                if _, ok := os.LookupEnv("KUBERNETES_SERVICE_PORT"); ok {
                    return true
                }
            }
        }
    }
    return false
}
```

---

## Command-Line Interface

### Command Hierarchy

```
prysm
├── local-producer
│   ├── ops-log
│   ├── disk-health-metrics
│   ├── kernel-metrics
│   └── resource-usage
├── remote-producer
│   ├── bucket-notify
│   ├── quota-usage-monitor
│   └── radosgw-usage
└── consumer
    └── quota-usage
```

### Example: Operations Log Command

**File**: `pkg/commands/producer_ops_log.go`

```go
var opsLogCmd = &cobra.Command{
    Use:   "ops-log",
    Short: "Start the operations log producer",
    Run: func(cmd *cobra.Command, args []string) {
        // 1. Load configuration from flags and environment
        cfg := opslog.LoadConfig()

        // 2. Create producer instance
        producer := opslog.New(cfg)

        // 3. Set up context with signal handling
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // Handle OS signals for graceful shutdown
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        go func() {
            <-sigChan
            log.Info().Msg("Shutdown signal received")
            cancel()
        }()

        // 4. Start the producer
        if err := producer.Start(ctx); err != nil {
            log.Fatal().Err(err).Msg("Failed to start producer")
        }
    },
}

func init() {
    // Define flags specific to ops-log
    opsLogCmd.Flags().String("log-file", "", "Path to ops log file")
    opsLogCmd.Flags().Int("prometheus-port", 8080, "Prometheus port")
    opsLogCmd.Flags().Bool("track-everything", false, "Enable all metrics")
    // ... more flags
}
```

**How it works**:
1. **Configuration Loading**: Combines CLI flags and environment variables
2. **Producer Initialization**: Creates producer with loaded config
3. **Signal Handling**: Graceful shutdown on SIGINT/SIGTERM
4. **Context Management**: Uses context for cancellation propagation
5. **Error Handling**: Fatal errors log and exit appropriately

---

## Producer Implementation

### Operations Log Producer Architecture

**File**: `pkg/producers/opslog/opslog.go`

#### Main Structure

```go
type OpsLogProducer struct {
    config           *Config
    logFile          *os.File
    watcher          *fsnotify.Watcher
    natsConn         *nats.Conn
    prometheusServer *http.Server
    auditor          Auditor

    // Dedicated storage maps for different metric types
    requestStorage   map[string]*RequestMetrics
    bytesStorage     map[string]*BytesMetrics
    errorStorage     map[string]*ErrorMetrics
    latencyStorage   map[string]*LatencyMetrics

    // Prometheus metrics
    metrics *PrometheusMetrics
}
```

#### Producer Lifecycle

```go
func (p *OpsLogProducer) Start(ctx context.Context) error {
    // 1. Initialize components
    if err := p.initLogFile(); err != nil {
        return err
    }

    if p.config.UseNATS {
        if err := p.connectNATS(); err != nil {
            return err
        }
    }

    if p.config.Prometheus {
        if err := p.startPrometheusServer(); err != nil {
            return err
        }
    }

    if p.config.AuditEnabled {
        p.initAuditor()
    }

    // 2. Set up file watcher
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    p.watcher = watcher

    if err := watcher.Add(p.config.LogFile); err != nil {
        return err
    }

    // 3. Start processing loop
    go p.processLoop(ctx)

    // 4. Wait for shutdown
    <-ctx.Done()
    return p.cleanup()
}
```

#### Log Processing Pipeline

```go
func (p *OpsLogProducer) processLoop(ctx context.Context) {
    // Use a buffered reader for efficiency
    reader := bufio.NewReader(p.logFile)

    for {
        select {
        case <-ctx.Done():
            return

        case event := <-p.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                // New data written to log file
                p.readNewLines(reader)
            }

        case err := <-p.watcher.Errors:
            log.Error().Err(err).Msg("File watcher error")
        }
    }
}

func (p *OpsLogProducer) readNewLines(reader *bufio.Reader) {
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err != io.EOF {
                log.Error().Err(err).Msg("Error reading line")
            }
            break
        }

        // Process the log entry
        p.processLogEntry(line)
    }
}

func (p *OpsLogProducer) processLogEntry(line string) {
    // 1. Parse JSON log entry
    var entry LogEntry
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
        log.Error().Err(err).Msg("Failed to parse log entry")
        return
    }

    // 2. Filter anonymous requests if configured
    if p.config.IgnoreAnonymousRequests && entry.User == "" {
        return
    }

    // 3. Update metrics in dedicated storage maps
    p.updateRequestMetrics(&entry)
    p.updateBytesMetrics(&entry)
    p.updateErrorMetrics(&entry)
    p.updateLatencyMetrics(&entry)

    // 4. Publish to NATS if enabled
    if p.config.UseNATS {
        p.publishToNATS(&entry)
    }

    // 5. Send to audit trail if enabled
    if p.config.AuditEnabled {
        p.publishAuditEvent(&entry)
    }
}
```

#### Dedicated Storage Architecture

**Problem**: High cardinality metrics can cause memory explosion

**Solution**: Separate storage maps per metric type with different aggregation levels

```go
// Example: Request metrics with multiple aggregation levels
func (p *OpsLogProducer) updateRequestMetrics(entry *LogEntry) {
    // Detailed metrics (full granularity)
    if p.config.TrackRequestsDetailed {
        key := fmt.Sprintf("%s|%s|%s|%s|%s",
            entry.Pod, entry.User, entry.Tenant, entry.Bucket, entry.Method)
        p.requestStorage[key].Increment()
    }

    // Per-user aggregation (all buckets combined)
    if p.config.TrackRequestsPerUser {
        key := fmt.Sprintf("%s|%s|%s", entry.Pod, entry.User, entry.Tenant)
        p.requestStorage[key].Increment()
    }

    // Per-bucket aggregation (all users combined)
    if p.config.TrackRequestsPerBucket {
        key := fmt.Sprintf("%s|%s|%s", entry.Pod, entry.Tenant, entry.Bucket)
        p.requestStorage[key].Increment()
    }

    // Per-tenant aggregation (all users and buckets)
    if p.config.TrackRequestsPerTenant {
        key := fmt.Sprintf("%s|%s", entry.Pod, entry.Tenant)
        p.requestStorage[key].Increment()
    }
}
```

**Benefits**:
- Only enabled metrics consume memory
- No runtime aggregation needed
- Each metric has optimal granularity
- Independent enable/disable per metric type

#### Prometheus Metrics Exposure

```go
type PrometheusMetrics struct {
    // Counters
    totalRequests       *prometheus.CounterVec
    totalRequestsUser   *prometheus.CounterVec
    totalRequestsBucket *prometheus.CounterVec
    bytesSent          *prometheus.CounterVec
    bytesReceived      *prometheus.CounterVec
    errors             *prometheus.CounterVec

    // Histograms for latency
    requestDuration         *prometheus.HistogramVec
    requestDurationPerUser  *prometheus.HistogramVec
    requestDurationPerMethod *prometheus.HistogramVec

    // Gauges for current state
    requestsByIP *prometheus.GaugeVec
}

func (p *OpsLogProducer) initPrometheusMetrics() {
    p.metrics = &PrometheusMetrics{
        totalRequests: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "radosgw_total_requests",
                Help: "Total number of requests with full detail",
            },
            []string{"pod", "user", "tenant", "bucket", "method", "http_status"},
        ),
        // ... more metrics
    }

    // Register all metrics with Prometheus
    prometheus.MustRegister(
        p.metrics.totalRequests,
        p.metrics.bytesSent,
        // ... more
    )
}

func (p *OpsLogProducer) startPrometheusServer() error {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())

    p.prometheusServer = &http.Server{
        Addr:    fmt.Sprintf(":%d", p.config.PrometheusPort),
        Handler: mux,
    }

    go func() {
        if err := p.prometheusServer.ListenAndServe(); err != nil {
            log.Error().Err(err).Msg("Prometheus server error")
        }
    }()

    return nil
}
```

### Disk Health Metrics Producer

**File**: `pkg/producers/diskhealthmetrics/diskhealthmetrics.go`

#### Main Structure

```go
type DiskHealthProducer struct {
    config    *Config
    natsConn  *nats.Conn
    metrics   *PrometheusMetrics
    osdMapper *OSDMapper  // Maps devices to Ceph OSD IDs
}

func (p *DiskHealthProducer) Start(ctx context.Context) error {
    // Initialize OSD mapper if Ceph integration enabled
    if p.config.CephOSDBasePath != "" {
        p.osdMapper = NewOSDMapper(p.config.CephOSDBasePath)
    }

    // Start collection loop
    ticker := time.NewTicker(time.Duration(p.config.Interval) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            p.collectMetrics()
        }
    }
}
```

#### SMART Data Collection

```go
func (p *DiskHealthProducer) collectMetrics() {
    for _, disk := range p.config.Disks {
        // Execute smartctl command
        data, err := p.executeSMARTCTL(disk)
        if err != nil {
            log.Error().Err(err).Str("disk", disk).Msg("Failed to collect SMART data")
            continue
        }

        // Parse SMART output
        metrics, err := p.parseSMARTData(data)
        if err != nil {
            log.Error().Err(err).Str("disk", disk).Msg("Failed to parse SMART data")
            continue
        }

        // Normalize attributes across vendors
        normalized := p.normalizeAttributes(metrics)

        // Map to OSD ID if Ceph integration enabled
        if p.osdMapper != nil {
            osdID := p.osdMapper.GetOSDID(disk)
            normalized.OSDID = osdID
        }

        // Update Prometheus metrics
        p.updatePrometheusMetrics(disk, normalized)

        // Publish to NATS if configured
        if p.config.UseNATS {
            p.publishToNATS(disk, normalized)
        }

        // Check thresholds and generate alerts
        p.checkThresholds(disk, normalized)
    }
}
```

#### SMART Attribute Normalization

```go
func (p *DiskHealthProducer) normalizeAttributes(raw *SMARTData) *NormalizedMetrics {
    normalized := &NormalizedMetrics{}

    // Temperature normalization
    if temp, ok := raw.Attributes["temperature"]; ok {
        normalized.Temperature = temp
    } else if temp, ok := raw.Attributes["airflow_temperature"]; ok {
        normalized.Temperature = temp
    }

    // Reallocated sectors (different vendors use different attribute IDs)
    if realloc, ok := raw.Attributes["reallocated_sector_ct"]; ok {
        normalized.ReallocatedSectors = realloc
    } else if realloc, ok := raw.Attributes["reallocated_event_count"]; ok {
        normalized.ReallocatedSectors = realloc
    }

    // SSD life used (varies by manufacturer)
    if life, ok := raw.Attributes["media_wearout_indicator"]; ok {
        // Intel SSDs: 100 = new, 0 = worn out
        normalized.SSDLifeUsed = 100 - life
    } else if life, ok := raw.Attributes["wear_leveling_count"]; ok {
        // Samsung SSDs: 100 = new
        normalized.SSDLifeUsed = 100 - life
    } else if life, ok := raw.Attributes["ssd_life_left"]; ok {
        // Some vendors report life left
        normalized.SSDLifeUsed = 100 - life
    }

    // NVMe specific attributes
    if raw.DeviceType == "nvme" {
        normalized.CriticalWarning = raw.NVMe.CriticalWarning
        normalized.AvailableSpare = raw.NVMe.AvailableSpare
        normalized.VendorID = fmt.Sprintf("0x%X", raw.NVMe.VendorID)
    }

    return normalized
}
```

#### Ceph OSD Mapping

```go
type OSDMapper struct {
    basePath string
    cache    map[string]string  // device -> OSD ID mapping
    mu       sync.RWMutex
}

func (m *OSDMapper) GetOSDID(device string) string {
    m.mu.RLock()
    if osdID, exists := m.cache[device]; exists {
        m.mu.RUnlock()
        return osdID
    }
    m.mu.RUnlock()

    // Discover OSD mapping
    osdID := m.discoverOSDID(device)

    m.mu.Lock()
    m.cache[device] = osdID
    m.mu.Unlock()

    return osdID
}

func (m *OSDMapper) discoverOSDID(device string) string {
    // Walk through Ceph OSD directories
    osdDirs, _ := filepath.Glob(filepath.Join(m.basePath, "osd*"))

    for _, osdDir := range osdDirs {
        // Read block device symlink
        blockPath := filepath.Join(osdDir, "block")
        target, err := os.Readlink(blockPath)
        if err != nil {
            continue
        }

        // Handle LVM logical volumes
        if strings.Contains(target, "/dev/mapper/") {
            // Resolve mapper to physical device
            physicalDev := m.resolveMapper(target)
            if physicalDev == device {
                return filepath.Base(osdDir)
            }
        } else if target == device {
            return filepath.Base(osdDir)
        }
    }

    return ""
}
```

---

## Consumer Implementation

### Quota Usage Consumer

**File**: `pkg/consumer/quotausageconsumer/quotausageconsumer.go`

```go
type QuotaUsageConsumer struct {
    config   *Config
    natsConn *nats.Conn
    metrics  *PrometheusMetrics
    alerts   chan Alert
}

func (c *QuotaUsageConsumer) Start(ctx context.Context) error {
    // Connect to NATS
    if err := c.connectNATS(); err != nil {
        return err
    }

    // Subscribe to quota usage subject
    sub, err := c.natsConn.Subscribe(c.config.NATSSubject, c.handleMessage)
    if err != nil {
        return err
    }
    defer sub.Unsubscribe()

    // Start Prometheus server
    if c.config.Prometheus {
        go c.startPrometheusServer()
    }

    // Start alert handler
    go c.handleAlerts(ctx)

    // Wait for shutdown
    <-ctx.Done()
    return nil
}

func (c *QuotaUsageConsumer) handleMessage(msg *nats.Msg) {
    // Parse quota usage message
    var usage QuotaUsage
    if err := json.Unmarshal(msg.Data, &usage); err != nil {
        log.Error().Err(err).Msg("Failed to parse quota message")
        return
    }

    // Update Prometheus metrics
    c.updateMetrics(&usage)

    // Check thresholds and generate alerts
    if c.isOverThreshold(&usage) {
        c.alerts <- Alert{
            Level:   "warning",
            User:    usage.User,
            Message: fmt.Sprintf("Quota at %d%%", usage.Percentage),
        }
    }
}
```

---

## Data Flow and Processing

### End-to-End Example: S3 Request Processing

```
1. User makes S3 request
   └─> RadosGW processes request

2. RadosGW writes to ops-log.log
   └─> {"time":"...","operation":"get_obj","user":"alice",...}

3. fsnotify detects file write
   └─> Prysm ops-log producer wakes up

4. Read and parse new log line
   └─> Unmarshal JSON into LogEntry struct

5. Process entry through pipeline
   ├─> Update dedicated storage maps
   │   ├─> Request counters
   │   ├─> Bytes transferred
   │   ├─> Error counters
   │   └─> Latency histograms
   │
   ├─> Update Prometheus metrics (direct mapping from storage)
   │
   ├─> Publish to NATS subject (if enabled)
   │   └─> "rgw.s3.ops" -> JSON payload
   │
   └─> Send to RabbitMQ audit trail (if enabled)
       └─> Convert to CADF format
       └─> Publish to "keystone.notifications.info"

6. Prometheus scrapes metrics
   └─> curl localhost:8080/metrics
   └─> Returns all current metric values

7. Grafana visualizes in dashboards
   └─> Queries Prometheus for metrics
   └─> Renders graphs and gauges

8. Alertmanager evaluates rules
   └─> Checks if error rate > threshold
   └─> Sends alert to PagerDuty/Slack
```

---

## Configuration Management

### Layered Configuration System

Prysm uses a multi-layered configuration approach:

```
Priority (highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values
```

### Example: Loading Configuration

```go
type Config struct {
    LogFile          string
    PrometheusPort   int
    UseNATS          bool
    NATSSubject      string
    TrackEverything  bool
    // ... more fields
}

func LoadConfig() *Config {
    cfg := &Config{
        // Default values
        PrometheusPort: 8080,
        NATSSubject:    "rgw.s3.ops",
    }

    // Load from environment variables
    if logFile := os.Getenv("LOG_FILE_PATH"); logFile != "" {
        cfg.LogFile = logFile
    }

    if port := os.Getenv("PROMETHEUS_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            cfg.PrometheusPort = p
        }
    }

    // Load from config file (if specified)
    if viper.ConfigFileUsed() != "" {
        viper.Unmarshal(cfg)
    }

    // Command-line flags override everything
    if flag := viper.GetString("log-file"); flag != "" {
        cfg.LogFile = flag
    }

    return cfg
}
```

---

## Messaging Layer

### NATS Integration

```go
func (p *Producer) connectNATS() error {
    // Parse NATS URL
    natsURL := p.config.NATSURL

    // Set up connection options
    opts := []nats.Option{
        nats.Name("prysm-producer"),
        nats.MaxReconnects(-1),  // Infinite reconnects
        nats.ReconnectWait(2 * time.Second),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            log.Warn().Err(err).Msg("NATS disconnected")
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            log.Info().Msg("NATS reconnected")
        }),
    }

    // Connect
    nc, err := nats.Connect(natsURL, opts...)
    if err != nil {
        return err
    }

    p.natsConn = nc
    log.Info().Str("url", natsURL).Msg("Connected to NATS")
    return nil
}

func (p *Producer) publishToNATS(entry *LogEntry) error {
    // Serialize to JSON
    data, err := json.Marshal(entry)
    if err != nil {
        return err
    }

    // Publish (fire-and-forget, non-blocking)
    return p.natsConn.Publish(p.config.NATSSubject, data)
}
```

### NATS Streaming (JetStream)

For guaranteed delivery:

```go
func (p *Producer) publishToJetStream(entry *LogEntry) error {
    js, err := p.natsConn.JetStream()
    if err != nil {
        return err
    }

    data, _ := json.Marshal(entry)

    // Publish with acknowledgment
    _, err = js.Publish(p.config.NATSSubject, data)
    return err
}
```

---

## Metrics Collection

### Prometheus Counter Pattern

```go
// Define metric
totalRequests := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "radosgw_total_requests",
        Help: "Total requests processed",
    },
    []string{"tenant", "bucket", "method"},
)

// Register with Prometheus
prometheus.MustRegister(totalRequests)

// Increment counter
totalRequests.WithLabelValues(
    entry.Tenant,
    entry.Bucket,
    entry.Method,
).Inc()
```

### Prometheus Histogram Pattern

```go
// Define histogram with custom buckets
requestDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "radosgw_requests_duration",
        Help: "Request duration in seconds",
        Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0, 10.0},
    },
    []string{"method"},
)

prometheus.MustRegister(requestDuration)

// Observe duration
duration := entry.TotalTime / 1000.0  // Convert ms to seconds
requestDuration.WithLabelValues(entry.Method).Observe(duration)
```

### Zero-Value Error Metrics

```go
// Problem: Error metrics disappear when no errors occur
// Solution: Always maintain the metric with value 0

func (p *Producer) ensureErrorMetric(labels ...string) {
    // Check if metric exists in storage
    key := strings.Join(labels, "|")
    if _, exists := p.errorStorage[key]; !exists {
        // Create with zero value
        p.errorStorage[key] = &ErrorMetrics{Count: 0}

        // Initialize Prometheus metric
        p.metrics.errors.WithLabelValues(labels...).Add(0)
    }
}

func (p *Producer) updateErrorMetrics(entry *LogEntry) {
    if entry.HTTPStatus >= 400 {
        // Error occurred
        p.metrics.errors.WithLabelValues(
            entry.Tenant,
            entry.Bucket,
            strconv.Itoa(entry.HTTPStatus),
        ).Inc()
    } else {
        // No error, but ensure metric exists with 0
        p.ensureErrorMetric(
            entry.Tenant,
            entry.Bucket,
            strconv.Itoa(entry.HTTPStatus),
        )
    }
}
```

---

## Key Design Patterns

### 1. Context-Based Cancellation

```go
func (p *Producer) Start(ctx context.Context) error {
    // Create child context for this component
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // Start background goroutines
    go p.processLoop(ctx)
    go p.metricsPublisher(ctx)

    // Wait for cancellation
    <-ctx.Done()

    // Cleanup
    return p.cleanup()
}
```

### 2. Graceful Shutdown

```go
func (p *Producer) cleanup() error {
    log.Info().Msg("Shutting down gracefully...")

    // Close file watcher
    if p.watcher != nil {
        p.watcher.Close()
    }

    // Close NATS connection
    if p.natsConn != nil {
        p.natsConn.Drain()  // Flush pending messages
        p.natsConn.Close()
    }

    // Shutdown Prometheus server
    if p.prometheusServer != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        p.prometheusServer.Shutdown(ctx)
    }

    // Close log file
    if p.logFile != nil {
        p.logFile.Close()
    }

    log.Info().Msg("Shutdown complete")
    return nil
}
```

### 3. Error Handling with Logging

```go
func (p *Producer) processLogEntry(line string) {
    var entry LogEntry
    if err := json.Unmarshal([]byte(line), &entry); err != nil {
        // Log error but continue processing
        log.Error().
            Err(err).
            Str("line", line).
            Msg("Failed to parse log entry")
        return
    }

    // Continue processing...
}
```

### 4. Concurrent Safe Operations

```go
type Storage struct {
    data map[string]int
    mu   sync.RWMutex
}

func (s *Storage) Increment(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key]++
}

func (s *Storage) Get(key string) int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.data[key]
}
```

### 5. Configuration Injection

```go
type Producer struct {
    config *Config
}

func New(cfg *Config) *Producer {
    return &Producer{
        config: cfg,
    }
}

// Allows for easy testing with mock configs
func TestProducer(t *testing.T) {
    cfg := &Config{
        UseNATS: false,  // Disable external dependencies for testing
        Prometheus: false,
    }
    producer := New(cfg)
    // Test producer...
}
```

---

## Summary

Prysm's codebase follows clean architecture principles:

1. **Separation of Concerns**: Commands, producers, consumers are independent
2. **Dependency Injection**: Configuration passed explicitly
3. **Context-Based Cancellation**: Proper goroutine lifecycle management
4. **Graceful Shutdown**: Clean resource cleanup
5. **Modular Design**: Easy to add new producers/consumers
6. **Observable**: Built-in logging and metrics for self-monitoring
7. **Testable**: Clear interfaces and dependency injection

The code is designed to be:
- **Maintainable**: Clear structure and naming
- **Extensible**: Easy to add new functionality
- **Reliable**: Proper error handling and recovery
- **Performant**: Efficient data structures and minimal allocations
- **Production-Ready**: Comprehensive logging, metrics, and health checks
