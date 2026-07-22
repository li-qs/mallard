# Mallard

[![Go Version](https://img.shields.io/badge/Go-1.26.2-blue.svg)](https://go.dev/)

这是一个轻量级的链路追踪工具，基于采样和分布式追踪，生成清晰的调用树。

_Mallard —— 动画片《鸭子侦探》 中的名侦探 梅小姐。_

## 快速开始

```bash
# 启动 Collector（默认 :8090）
go run main.go

# 自定义端口
go run main.go 9090
```

打开浏览器访问 `http://localhost:8090/ui/` 查看 UI。Demo：https://static.planxt.site/mallard/

客户端接入请使用 [mallard-tracer](https://github.com/li-qs/mallard-tracer)。

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

| 字段       | 类型   | 说明                   |
| ---------- | ------ | ---------------------- |
| trace_id   | string | 调用链 ID              |
| span_id    | string | Span ID                |
| parent_id  | string | 父 Span ID，根节点为空 |
| service    | string | 服务名                 |
| operation  | string | 操作名                 |
| start_time | int64  | 开始时间（纳秒）       |
| duration   | int64  | 耗时（纳秒）           |
| status     | int    | HTTP 状态码            |
| error      | string | 错误信息（可选）       |

### 查询 Trace

```
GET /trace/{trace_id}
```

返回该 Trace 下所有 Span 的 JSON 数组。

## UI 功能

Web UI 截图:

![alt text](https://github.com/user-attachments/assets/186e416a-b644-408e-bdc0-bb8fc75867a5)

- 调用链
  - 调用树：父子 Span 之间的调用关系，出现 ❌ 表示调用错误；鼠标 Hover 出现复制图标 📋，点击可复制当前 Span 详情。
  - 瀑布图：每个 Span 的发生时间、耗时等；鼠标 Hover 展示 Span 详情。
- 重复调用
  - 调用次数：该 Span 被重复调用的次数；出现重复调用时需要关注业务设计是否合理。
  - 调用关系：高亮展示重复调用的 Span；鼠标点击任意 Span 可复制 Span 详情。

## TODO

- [ ] 数据持久化：MongoDB 存储扁平化数据；
- [ ] 智能采样策略：头尾采样、错误采样，样本率、强制上报；
- [ ] 批量上报：本地缓冲队列-批量上报（时间触发、容量触发）；

## 项目结构

```
mallard/
├── main.go           # 入口，启动 HTTP 服务
├── collector/        # 数据采集与存储
├── html/
│   └── index.html    # Web UI（瀑布图 + 重复调用 + 关系图）
└── go.mod
```

## License

[MIT](./LICENSE)
