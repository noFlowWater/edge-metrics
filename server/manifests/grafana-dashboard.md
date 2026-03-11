# Edge Devices Power & Energy Monitoring Dashboard

Grafana dashboard for real-time power metric visualization and workload energy analysis across Jetson edge devices (Xavier, Nano, Orin).

## Per-Device Power Metric Structure

| Device | Total Power Metric | Description |
|--------|-------------------|-------------|
| **Xavier** | `jetson_power_vdd_in_watts` | Main power input (single sensor) |
| **Nano** | `jetson_power_pom_5v_in_watts` | 5V input power (single sensor) |
| **Orin** | Sum required | Individual rails only |

### Orin Power Calculation

Orin lacks a single total power sensor, so 3 channels must be summed:

```promql
jetson_power_vdd_gpu_soc_watts    # GPU + SoC power
+ jetson_power_vdd_cpu_cv_watts   # CPU + Computer Vision power
+ jetson_power_vin_sys_5v0_watts  # System 5V power
```

---

## Energy Calculation

### Formula

```
Energy (Wh) = Average Power (W) x Time (h)
```

### PromQL Implementation

```promql
avg_over_time(power_metric[$__range]) * $__range_s / 3600
```

| Variable | Description |
|----------|-------------|
| `$__range` | Grafana selected time range (e.g., 15m, 1h) |
| `$__range_s` | Time range in seconds |
| `/ 3600` | Convert seconds to hours (for Wh) |

### Example

```
Time range: 15 min = 900 sec = 0.25 h
Average power: 5W
Energy: 5W x 0.25h = 1.25 Wh
```

### Notes

- `increase()` is for **Counter** metrics only
- Power metrics are **Gauge** type, so use `avg_over_time()`

---

## Statistical Functions

| Function | Purpose | Example |
|----------|---------|---------|
| `avg_over_time(metric[$__range])` | Average over time range | Average power |
| `max_over_time(metric[$__range])` | Maximum over time range | Peak power |
| `min_over_time(metric[$__range])` | Minimum over time range | Idle power |

---

## Panel Query Details

### 1. Real-Time Power Metrics (Time Series)

```promql
jetson_power_vdd_in_watts{device_type=~"$device_type",hostname=~"$hostname"}
```

- **interval**: `1s` (1-second resolution)
- **Purpose**: Real-time power monitoring

### 2. Total Energy Usage (Stat)

**Xavier:**
```promql
sum(avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])) * $__range_s / 3600
```

**Nano:**
```promql
sum(avg_over_time(jetson_power_pom_5v_in_watts{device_type="jetson_nano"}[$__range])) * $__range_s / 3600
```

**Orin (3-channel sum):**
```promql
sum(avg_over_time(jetson_power_vdd_gpu_soc_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
+ sum(avg_over_time(jetson_power_vdd_cpu_cv_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
+ sum(avg_over_time(jetson_power_vin_sys_5v0_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
```

### 3. Average / Max / Min Power (Stat)

**Average:**
```promql
avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

**Max:**
```promql
max_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

**Min:**
```promql
min_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

### 4. Per-Device Energy Ratio (Pie Chart)

```promql
sum by (hostname) (avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])) * $__range_s / 3600
```

### 5. Per-Device Statistics Table

Multiple device types joined with `or` in a single table:

```promql
label_replace(avg_over_time(jetson_power_vdd_in_watts{...}[$__range]), "metric", "Avg Power (W)", "", "")
or label_replace(avg_over_time(jetson_power_pom_5v_in_watts{...}[$__range]), "metric", "Avg Power (W)", "", "")
or label_replace(avg_over_time(jetson_power_vin_sys_5v0_watts{...}[$__range]), "metric", "Avg Power (W)", "", "")
```

---

## Query Resolution Settings

| Setting | Value | Description |
|---------|-------|-------------|
| Query `interval` | `1s` | Grafana query resolution |
| ServiceMonitor `scrapeInterval` | `1s` | Prometheus scrape interval |

Grafana auto-calculates step based on time range. Set `interval: "1s"` for exact 1-second resolution.

---

## Dashboard Variables

| Variable | Query | Description |
|----------|-------|-------------|
| `$device_type` | `label_values({__name__=~"jetson_power_.*"}, device_type)` | Device type filter |
| `$hostname` | `label_values({__name__=~"jetson_power_.*", device_type=~"$device_type"}, hostname)` | Hostname filter |

---

## Import Instructions

1. Log in to Grafana
2. Go to `Dashboards` -> `Import`
3. Click `Upload JSON file`
4. Upload `grafana-dashboard.json`
5. Select Prometheus data source
6. Click `Import`

---

## Workload Energy Analysis

### Procedure

1. Record workload start time
2. Run workload
3. Record workload end time
4. Set Grafana Time Range (start to end)
5. Check statistics:
   - Total energy usage (Wh)
   - Average power (W)
   - Peak power (W)

### Energy Efficiency Comparison Example

| Workload | Duration | Avg Power | Total Energy | Efficiency |
|----------|----------|-----------|--------------|------------|
| Model A | 5 min | 3.8W | 0.32 Wh | Baseline |
| Model B | 5 min | 3.2W | 0.27 Wh | 16% savings |

---

## Related Files

- `grafana-dashboard.json`: Dashboard JSON definition
- `servicemonitor.yaml`: Prometheus ServiceMonitor config
- `prometheus-values.yaml`: Helm values (exclude node-exporter jetson metrics)
