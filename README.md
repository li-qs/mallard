# Mallard

[![Go Version](https://img.shields.io/badge/Go-1.26.2-blue.svg)](https://go.dev/)

Mallard 是一个轻量级的链路追踪工具，基于采样和分布式追踪，生成清晰的调用树。无论是定位慢调用、诊断异常，还是梳理服务依赖，Mallard 都能帮你快速找到"案发现场"。

🦆 致敬《鸭子侦探》中的梅小姐 —— 她用敏锐的观察力破解疑案，我们用精准的链路数据破解系统故障。

## 快速开始

```bash
# 启动 Collector（默认 :8090）
go run main.go

# 自定义端口
go run main.go 9090
```

打开浏览器访问 `http://localhost:8090/ui/` 查看 UI。

## API

### 上报 Span

```
POST /collect
Content-Type: application/json

[
  {
    "trace_id": "18c42782b4d78148c67c8",
    "span_id": "18c42782b4d798b8c7f38",
    "parent_id": "",
    "service": "w1",
    "operation": "GET /goods",
    "start_time": 1784594794678821000,
    "duration": 1410357703,
    "status": 200
  }
]
```

| 字段 | 类型 | 说明 |
|------|------|------|
| trace_id | string | 调用链 ID |
| span_id | string | Span ID |
| parent_id | string | 父 Span ID，根节点为空 |
| service | string | 服务名 |
| operation | string | 操作名 |
| start_time | int64 | 开始时间（纳秒） |
| duration | int64 | 耗时（纳秒） |
| status | int | HTTP 状态码 |
| error | string | 错误信息（可选） |

### 查询 Trace

```
GET /trace/{trace_id}
```

返回该 Trace 下所有 Span 的 JSON 数组。

## UI 功能

Web UI 截图:

![alt text](https://github.com/user-attachments/assets/186e416a-b644-408e-bdc0-bb8fc75867a5)

- 调用链瀑布图
- 重复调用分析
  - 调用关系图


## 项目结构

```
mallard/
├── main.go           # 入口，启动 HTTP 服务
├── collector/        # 数据采集与存储
├── html/
│   └── index.html    # Web UI（瀑布图 + 重复调用 + 关系图）
└── go.mod
```
