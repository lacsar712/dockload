# dockload

集装箱码头 **吊具称重流处理** 服务：接收载荷传感器读数，经皮重/量程标定与稳定窗口判定，生成有效称重事件供码头作业系统（TOS）消费。

## 功能

- **标定（calib）**：皮重 `tare` + 量程斜率 `span`，并发安全存储
- **稳定窗口（stable）**：最近 N 个样本的总体标准差低于阈值视为稳定
- **称重状态机（weigh）**：`Idle → Collecting → Stable → Published → Idle`
- **传感器接入（ingest）**：接收原始计数 `rawCounts`
- **事件发布（publish）**：稳定后生成 `WeighEvent` 并发布
- **嵌入式 Web 操作台**：查看标定与最近称重记录

## 快速开始

```bash
go build -o dockload ./cmd/dockload
./dockload -addr :8080
```

浏览器访问 `http://localhost:8080/` 打开操作台。

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/sensors/raw` | 上报传感器原始计数 |
| PUT | `/v1/calib/{hookID}` | 设置吊具标定 |
| GET | `/v1/calib/{hookID}` | 查询吊具标定 |
| GET | `/v1/weighs/recent` | 最近称重事件 |
| GET | `/` | Web 操作台 |

### 上报原始读数

```json
POST /v1/sensors/raw
{
  "hookID": "H7",
  "rawCounts": 1500,
  "timestamp": "2026-08-20T09:00:00Z"
}
```

### 设置标定

```json
PUT /v1/calib/H7
{
  "tare": 1000,
  "span": 0.05,
  "maxLoadKg": 45000
}
```

## 配置

环境变量（均可选）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `DOCKLOAD_ADDR` | `:8080` | 监听地址 |
| `DOCKLOAD_WINDOW_SIZE` | `10` | 稳定窗口样本数 |
| `DOCKLOAD_STABLE_EPS` | `5.0` | 稳定标准差阈值（counts） |
| `DOCKLOAD_LOAD_CHANGE` | `50.0` | 稳定后载荷骤变阈值（counts） |
| `DOCKLOAD_MAX_RECENT` | `100` | 保留最近事件数 |

## 业务规则

1. 净重计算：`netKg = (rawCounts - tare) * span`
2. `span <= 0` 或未设置标定 → 拒绝，原因 `NOT_CALIBRATED`
3. 净重超过 `maxLoadKg` → 拒绝，原因 `OVER_RANGE`
4. 窗口未满或标准差过高 → 拒绝，原因 `NOT_STABLE`

## 开发

```bash
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go test ./... -count=1
```

## 许可证

MIT
