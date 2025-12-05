---
title: Stress Testing
tags:
  - guide
---

# Stress Testing

Load test APIs with concurrent requests and track performance metrics.

## Overview

Stress testing feature enables:

- Concurrent request execution
- Configurable load profiles
- Real-time progress tracking
- Performance metrics (latency, throughput)
- Historical test results

## Creating a Stress Test

Press `S` in TUI mode to open stress test viewer, then `n` for new test.

### Configuration Fields

**Config Name** (optional)
Unique identifier for saving/reusing configuration.

**Request File**
Path to `.http` file. Pre-filled from current file.

**Concurrent Connections**
Number of parallel workers (1-1000). Default: 10.

**Total Requests**
Total number of requests to send. Default: 100.

**Ramp-Up Duration** (seconds)
Time to gradually increase load. 0 = no ramp-up.

**Test Duration** (seconds)
Maximum test duration. 0 = unlimited (stops when all requests complete).

### Example Configuration

```text
┌─ Stress Test Configuration ──────────────┐
│                                          │
│ Config Name:           api-load-test     │
│ Request File:          /path/to/api.http │
│ Request Name:          GET users         │
│ Concurrent Connections: 50               │
│ Total Requests:        1000              │
│ Ramp-Up Duration (sec): 10               │
│ Test Duration (sec):   60                │
│                                          │
│ Ctrl+S: Save & Start | ESC: Cancel       │
└──────────────────────────────────────────┘
```

### Keyboard Shortcuts (Config Mode)

| Key            | Action                     |
| -------------- | -------------------------- |
| `↑/↓` or `j/k` | Navigate fields            |
| `Enter`        | Edit field / move to next  |
| `Ctrl+S`       | Save config and start test |
| `Ctrl+L`       | Load saved configuration   |
| `ESC`          | Cancel                     |

## Running a Test

After pressing `Ctrl+S`, test starts immediately.

### Progress View

```text
┌─ Stress Test - Running ─────────────────┐
│                                         │
│ Progress                                │
│ 450/1000 requests (45.0%)               │
│ ███████████░░░░░░░░░░░░░░               │
│ Elapsed: 15.2s                          │
│                                         │
│ Statistics                              │
│ Success:    445       Max:        456ms │
│ Errors:     5         P50:        125ms │
│ Avg:        132ms     P95:        289ms │
│ Min:        45ms      P99:        378ms │
│                                         │
│ Requests/sec: 29.61                     │
│                                         │
│ ESC: Stop test | r: View results        │
└─────────────────────────────────────────┘
```

### Live Metrics

- **Progress Bar**: Visual completion percentage
- **Success/Errors**: Count of successful (2xx-3xx) vs failed requests
- **Latency**: avg, min, max, P50 (median), P95, P99 percentiles
- **Throughput**: Requests per second
- **Elapsed Time**: Duration since test start

### Stopping a Test

Press `ESC` to stop test early. Partial results will be saved.

## Viewing Results

Press `S` to open stress test results (split view).

### Results View

```text
┌─────────────────────────┬─────────────────────────┐
│ Test Runs               │ Details                 │
│                         │                         │
│ > ✓ api-load-test       │ api-load-test           │
│   2025-12-04 10:30      │                         │
│   1000 reqs | 132ms avg │ Status                  │
│                         │ Status:     completed   │
│   ✓ auth-test           │ Started:    10:30:15    │
│   2025-12-04 09:15      │ Completed:  10:31:45    │
│   500 reqs | 89ms avg   │ Duration:   1m 30s      │
│                         │                         │
│   ○ quick-test          │ Requests                │
│   2025-12-04 08:00      │ Sent:       1000        │
│   100 reqs | 45ms avg   │ Completed:  1000        │
│                         │ Success:    995         │
│                         │ Errors:     5           │
│                         │ Success:    99.5%       │
│                         │                         │
│                         │ Latency                 │
│                         │ Average:    132ms       │
│                         │ Min:        45ms        │
│                         │ Max:        456ms       │
│                         │ P50:        125ms       │
│                         │ P95:        289ms       │
│                         │ P99:        378ms       │
└─────────────────────────┴─────────────────────────┘
```

### Status Icons

- `✓` Completed successfully
- `✗` Failed (error during execution)
- `○` Cancelled (stopped early)
- `◐` Running (in progress)

### Keyboard Shortcuts (Results Mode)

| Key            | Action                      |
| -------------- | --------------------------- |
| `↑/↓` or `j/k` | Navigate test runs (list)   |
| `TAB`          | Switch focus (list/details) |
| `d`            | Delete selected run         |
| `n`            | Create new stress test      |
| `ESC` or `q`   | Close viewer                |

## Saved Configurations

Named configurations are stored for reuse.

### Saving a Config

1. Enter unique name in "Config Name" field
2. Press `Ctrl+S` to save and start
3. Config is automatically saved to database

### Loading a Config

1. In config mode, press `Ctrl+L`
2. Select saved configuration
3. Press `Enter` to load
4. Modify as needed and press `Ctrl+S`

### Managing Configs

Navigate list with `↑/↓`, press `d` to delete unwanted configs.

## Ramp-Up Strategy

Gradually increases load to avoid overwhelming target server.

### How It Works

Requests are distributed evenly across ramp-up duration:

```text
Ramp-up: 30 seconds
Total requests: 300
Distribution: 10 requests/second

0s ─────────────────────── 30s
│ ░░░░░░░░░░░░░░░░░░░░░░ │
Start                    Peak

Concurrent: 10 workers
Each starts 1 request every 3 seconds
```

### Use Cases

**No Ramp-Up** (0 seconds)

- Quick tests
- Known stable endpoints
- Max performance testing

**Gradual Ramp-Up** (10-60 seconds)

- Production-like load testing
- Identify breaking points
- Avoid connection flooding

## Performance Metrics

### Latency Percentiles

**P50 (Median)**
50% of requests faster than this value.

**P95**
95% of requests faster than this value. Good indicator of "worst typical case".

**P99**
99% of requests faster than this value. Catches outliers.

### Example Analysis

```text
Avg: 125ms  P50: 120ms  P95: 250ms  P99: 450ms
```

**Interpretation:**

- Typical request: ~120ms
- Most requests: < 250ms
- 1% outliers: up to 450ms

Investigate if P95/P99 significantly higher than avg.

## Data Storage

Stress test data stored in SQLite database:

```text
~/.restcli/restcli.db
```

**Tables:**

- `stress_test_configs`: Saved configurations
- `stress_test_runs`: Test execution metadata
- `stress_test_metrics`: Per-request timing data

**Storage Considerations:**

- 1000 requests ≈ 200KB data
- 10,000 requests ≈ 2MB data
- No automatic cleanup (manual delete via UI)

## Best Practices

### Target Server

**Production**

- Start with low concurrent connections (5-10)
- Use ramp-up to avoid spikes
- Monitor server resources
- Coordinate with ops team

**Development/Staging**

- Higher concurrency acceptable
- Test failure scenarios
- Verify rate limiting

### Request Selection

**Best Results:**

- Idempotent endpoints (GET, HEAD)
- Read operations
- Cacheable responses

**Caution:**

- POST/PUT/DELETE (may create data)
- Non-idempotent operations
- Rate-limited endpoints

### Configuration Guidelines

```text
Light Load:    10 connections, 100 requests
Medium Load:   50 connections, 1000 requests
Heavy Load:    100 connections, 10000 requests
Spike Test:    200 connections, 1000 requests, 0s ramp-up
Sustained:     50 connections, unlimited, 300s duration
```
