# Edge Devices Power & Energy Monitoring Dashboard

Jetson 엣지 디바이스(Xavier, Nano, Orin)의 전력 메트릭을 실시간으로 시각화하고, 워크로드별 에너지 사용량을 분석하기 위한 Grafana 대시보드입니다.

## 디바이스별 전력 메트릭 구조

| 디바이스 | 총 전력 메트릭 | 설명 |
|---------|--------------|------|
| **Xavier** | `jetson_power_vdd_in_watts` | 메인 전원 입력 (단일 센서) |
| **Nano** | `jetson_power_pom_5v_in_watts` | 5V 입력 전원 (단일 센서) |
| **Orin** | 합산 필요 | 개별 레일만 존재 |

### Orin 전력 계산

Orin은 총 전력을 측정하는 단일 센서가 없어 3개 채널을 합산해야 합니다:

```promql
jetson_power_vdd_gpu_soc_watts    # GPU + SoC 전력
+ jetson_power_vdd_cpu_cv_watts   # CPU + Computer Vision 전력
+ jetson_power_vin_sys_5v0_watts  # 시스템 5V 전력
```

---

## 에너지 계산 수식

### 기본 공식

```
에너지 (Wh) = 평균 전력 (W) × 시간 (h)
```

### PromQL 구현

```promql
avg_over_time(power_metric[$__range]) * $__range_s / 3600
```

| 변수 | 설명 |
|-----|------|
| `$__range` | Grafana 선택 시간 범위 (예: 15m, 1h) |
| `$__range_s` | 시간 범위를 초 단위로 변환 |
| `/ 3600` | 초 → 시간 변환 (Wh 단위) |

### 계산 예시

```
시간 범위: 15분 = 900초 = 0.25시간
평균 전력: 5W
에너지: 5W × 0.25h = 1.25 Wh
```

### 주의사항

- `increase()` 함수는 **Counter** 메트릭 전용
- 전력 메트릭은 **Gauge** 타입이므로 `avg_over_time()` 사용

---

## 통계 함수

| 함수 | 용도 | 사용 예시 |
|-----|------|----------|
| `avg_over_time(metric[$__range])` | 시간 범위 내 평균값 | 평균 전력 계산 |
| `max_over_time(metric[$__range])` | 시간 범위 내 최대값 | 피크 전력 감지 |
| `min_over_time(metric[$__range])` | 시간 범위 내 최소값 | 유휴 전력 확인 |

---

## 패널별 쿼리 상세

### 1. 실시간 전력 메트릭 (Time Series)

```promql
jetson_power_vdd_in_watts{device_type=~"$device_type",hostname=~"$hostname"}
```

- **interval**: `1s` (1초 해상도)
- **용도**: 전력 변화 실시간 모니터링

### 2. 총 에너지 사용량 (Stat)

**Xavier:**
```promql
sum(avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])) * $__range_s / 3600
```

**Nano:**
```promql
sum(avg_over_time(jetson_power_pom_5v_in_watts{device_type="jetson_nano"}[$__range])) * $__range_s / 3600
```

**Orin (3채널 합산):**
```promql
sum(avg_over_time(jetson_power_vdd_gpu_soc_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
+ sum(avg_over_time(jetson_power_vdd_cpu_cv_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
+ sum(avg_over_time(jetson_power_vin_sys_5v0_watts{device_type="jetson_orin"}[$__range])) * $__range_s / 3600
```

### 3. 평균/최대/최소 전력 (Stat)

**평균 전력:**
```promql
avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

**최대 전력:**
```promql
max_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

**최소 전력:**
```promql
min_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])
```

### 4. 디바이스별 에너지 비율 (Pie Chart)

```promql
sum by (hostname) (avg_over_time(jetson_power_vdd_in_watts{device_type="jetson_xavier"}[$__range])) * $__range_s / 3600
```

### 5. 디바이스별 통계 테이블 (Table)

여러 디바이스 타입을 `or`로 연결하여 단일 테이블에 표시:

```promql
label_replace(avg_over_time(jetson_power_vdd_in_watts{...}[$__range]), "metric", "평균 전력 (W)", "", "")
or label_replace(avg_over_time(jetson_power_pom_5v_in_watts{...}[$__range]), "metric", "평균 전력 (W)", "", "")
or label_replace(avg_over_time(jetson_power_vin_sys_5v0_watts{...}[$__range]), "metric", "평균 전력 (W)", "", "")
```

---

## 쿼리 해상도 설정

| 설정 | 값 | 설명 |
|-----|---|------|
| Query `interval` | `1s` | Grafana 쿼리 해상도 |
| ServiceMonitor `scrapeInterval` | `1s` | Prometheus 수집 주기 |

Grafana는 시간 범위에 따라 자동으로 step을 계산하므로, 정확한 1초 해상도가 필요하면 `interval: "1s"` 명시 필요.

---

## 대시보드 변수

| 변수 | 쿼리 | 설명 |
|-----|------|------|
| `$device_type` | `label_values({__name__=~"jetson_power_.*"}, device_type)` | 디바이스 타입 필터 |
| `$hostname` | `label_values({__name__=~"jetson_power_.*", device_type=~"$device_type"}, hostname)` | 호스트명 필터 |

---

## Import 방법

1. Grafana 로그인
2. `Dashboards` → `Import` 클릭
3. `Upload JSON file` 선택
4. `grafana-dashboard.json` 업로드
5. Prometheus 데이터소스 선택
6. `Import` 클릭

---

## 워크로드 에너지 분석

### 분석 절차

1. **워크로드 시작 시간 기록**
2. **워크로드 실행**
3. **워크로드 종료 시간 기록**
4. **Grafana Time Range 설정** (시작~종료 시간)
5. **통계 확인**:
   - 총 에너지 사용량 (Wh)
   - 평균 전력 (W)
   - 최대 전력 (W)

### 에너지 효율 비교 예시

| 워크로드 | 시간 | 평균 전력 | 총 에너지 | 효율 |
|---------|-----|----------|----------|------|
| Model A | 5분 | 3.8W | 0.32 Wh | 기준 |
| Model B | 5분 | 3.2W | 0.27 Wh | 16% 절감 |

---

## 관련 파일

- `grafana-dashboard.json`: 대시보드 JSON 정의
- `servicemonitor.yaml`: Prometheus ServiceMonitor 설정
- `prometheus-values.yaml`: Helm values (node-exporter jetson 메트릭 제외)
