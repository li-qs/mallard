# TODO — 链路追踪系统（mallard）

> 基于 Echo v5 + MongoDB。记录待实现的功能与优化，按优先级排序。
> 已完成部分见 README「API」章节。

## 当前状态

已实现：
- 用户认证（登录/登出/刷新，JWT + Refresh Token，bcrypt）
- App 管理（增删改、secret 轮换、IP allow list、Basic 认证、本地 TTL 缓存 + SHA-256）
- Span 批量上报（App 认证，`InsertMany` 去重，`span_id` 唯一索引）
- Trace 查询（`GET /traces/:trace_id`，start_time 正序，含 `is_root` 根标记，M2 完成）
- Trace 搜索列表（`GET /traces`，聚合过滤 + 分页，M1 完成）
- 探针（`/health`、`/ready`）

---

## 一、查询/展示类（P0：核心价值）

### 1. Trace 搜索列表 `GET /traces` ✅ M1 已完成
- 支持过滤：`app_id`、`operation`、`status`、`trace_id`（前缀模糊）、`start_time_gt`/`start_time_lt`（纳秒）
- 分页：`page`/`page_size`（默认 1/10，上限 100），返回 `ListData<TraceSummary>`（app_ids、operation、start_time、duration、span_count、error_count、has_error）
- 实现：`$match` → `$group by trace_id`（`$min` 开始时间、`$max(start+duration)` 计算链路时长、`$first` 取根 span operation、`error_count` 统计 status≠0）→ 按 start_time 倒序 → 分页；`CountTraces` 单独 `$group+$count`
- 已加 `{start_time: -1}` 索引支持纯时间范围查询
- `status` 筛选语义：不传=全部，`1`=成功（`error_count==0`），`2`=错误（`error_count>0`），在 `$group` 后按 `error_count` 过滤；span 数据侧仍以 `status != 0` 判定错误

### 2. Trace 详情树 / 瀑布图 `GET /traces/:trace_id` ✅ M2 已完成（方案 B）
- 方案 B：保持平铺返回（按 start_time 正序），前端自行按 `parent_id` 组树
- 已在响应增加 `is_root` 字段（`parent_id` 为空即为根），前端可直接用
- 若前端需要嵌套树结构再评估方案 A（后端 `children` 字段）

### 3. 聚合统计 `GET /stats` ⭐
**用途**：各 service/operation 的 QPS、错误率、P50/P95/P99 延迟、慢接口 TopN。tracing 的核心价值。

- **接口拆分建议**：
  - `GET /stats/services`：服务维度汇总
  - `GET /stats/operations?app_id=&operation=`：单接口明细
  - `GET /stats/slow?top=10`：慢接口 TopN（按 P95 排序）
- **请求参数**：`app_id`、`operation`、`start_time_gt`/`start_time_lt`、`granularity`（minute/hour/day，时间粒度）
- **实现要点**：
  - 聚合管道：`$match` 时间范围 → `$group`（service/operation + 时间桶）→ 计算 count、error_count、延迟 `$percentile`（MongoDB 7.0+ 支持 `$percentile`，旧版本需 `$push` + 应用层算）
  - 延迟单位纳秒，响应转毫秒展示
  - 时间桶：`$dateTrunc`（reported_at 或 start_time 转 date 后截断）
- **注意**：start_time 是 int64 纳秒，需先换算成 Date 才能用日期聚合算子

### 4. 服务列表 / 拓扑（Service Map）`GET /services`、`GET /topology`
**用途**：跨服务调用关系图，快速定位依赖链问题。

- `GET /services`：所有上报过的服务/operation 去重列表
  - 实现：`$group` 按 app_id + operation 去重，或 distinct
- `GET /topology?start_time_gt=&start_time_lt=`：服务依赖边（谁调用谁）
  - 实现：按 `trace_id` 内，利用 parent/child 关系推断调用方/被调用方
  - 该功能依赖 trace 树的构建，建议在 #2 完成后再做

---

## 二、摄取/接入类（P1）

### 5. 单条 span 上报 `POST /span`
**用途**：调试、低流量场景，避免为单条建 batch 的浪费。
- 复用 `dto.Span`，`POST /spans` 改为同时接受 `{ "spans": [...] }` 或单条对象（或单独 `POST /span`）
- 响应同 `vo.SpanReport{accepted}`

### 6. OTLP 协议兼容（OpenTelemetry）⭐
**用途**：让标准 OpenTelemetry SDK（Java/Go/Python 等）直接上报，不用自研 SDK。业界主流接入形态。

- **HTTP/JSON**：`POST /v1/traces`，兼容 OTLP JSON（`ExportTraceServiceRequest`）
- **gRPC**：`POST /v1/traces`（HTTP/2 + protobuf），需引入 `go.opentelemetry.io/proto/otlp` 或自建 protobuf
- **实现**：新增一个 converter 层，把 OTLP `ResourceSpans → ScopeSpans → Span` 映射为 `model.Span`（注意映射：trace_id/span_id 是字节数组需转 hex，时间戳转纳秒，status 转 int，attributes 转 operation 或 error）
- **认证**：沿用 AppAuth，凭据放 gRPC metadata / HTTP Basic header
- **影响面**：新增 handler + converter + 依赖，改动较大，建议单独里程碑

### 7. 采样 / 限流 / 背压
**用途**：高并发上报时保护服务，控制存储成本。

- 服务端采样：按 app 维度配置采样率（`head sampling`），上报时抽样丢弃或标记
- 上报限流：`/spans` 接口按 IP 限流（参照 `/login` 的 RateLimiter 配置）
- 批量大小限制：单次上报 spans 数量上限（如 1000），超出拒绝或拆分
- 背压：上游过载时返回 429，客户端退避重试

---

## 三、管理/运维类（P1-P2）

### 8. App allow list 校验（已完成）
**现状**：`ip_allow_list` 字段已生效，`AppAuth` 中间件会校验 `c.RealIP()`：
- allow list 为空 → 不限制 IP（默认放行）
- 支持精确 IP 与 CIDR（如 `192.168.1.0/24`），校验不通过返回 403
- 补充项：`RealIP()` 依赖 `X-Forwarded-For`/`X-Real-IP`，生产环境须确保前置代理会覆盖这些头（防伪造）

### 9. 保留策略动态配置
**现状**：span TTL 写死在 `config.yaml` 的 `tracing_data.spans_expire_seconds`，改配置需重启。
- 目标：支持运行时按 app 调整保留天数，或独立配置接口
- 注意：MongoDB TTL 索引建好后改 `expireAfterSeconds` 需重建索引（`collMod`），动态调整涉及索引维护

### 10. Prometheus metrics
**用途**：服务自身监控（请求量、延迟、错误率、span 积压）。
- 引入 `prometheus/client_golang`，暴露 `/metrics`
- 指标建议：`http_requests_total`、`http_request_duration_seconds`、`spans_received_total`、`spans_rejected_total`
- Echo 有官方 `middleware.Prometheus` 插件可直接接入

### 11. 健康检查细化
**现状**：`/health` 恒 200，`/ready` 只 ping Mongo。
- 优化：`/ready` 同时校验 Mongo 可用 + 最近 span 写入是否正常（可选项）
- 增加慢查询/依赖超时时间可配置

---

## 四、技术债 / 工程质量（贯穿）

### 12. 单元测试
**现状**：0 个测试文件。
- service 层：App（secret 校验、缓存失效）、Login（token 生成/刷新/过期）、Span（去重逻辑）
- repository 层：可用 `mongo-driver` 的 `mtest`（mongodb mock server）
- handler 层：`echo` + `httptest`，Mock service
- 优先测：认证、去重、分页、聚合管道

### 13. 错误码与日志规范化
- `response.JsonError` 目前返回 HTTP 200 + body code，前端依赖文档约定；建议整理统一错误码表
- handler 错误日志已结构化（zap.Error），但个别地方仍 `err.Error()` 直接透出，需审查（避免泄漏内部信息）

### 14. 配置项梳理
- `config.yaml` 中 mongo 连接串含明文密码，正式环境务必用环境变量覆盖（已支持 `AutomaticEnv`）
- 可考虑引入 `.env.example` 模板
- `cookie_secure` 已配置化，确认生产开启

### 15. 分页/查询参数统一
- App 列表与 Trace 搜索的 `page`/`page_size` 解析逻辑重复，抽公共 helper（如 `utils.ParsePagination(c)`）
- 分页上限、默认值统一管理

### 16. 性能
- span 上报：考虑写缓冲/批量落地（worker pool + channel），减少高频单次写
- 大 trace 查询：限制单 trace span 数量上限（如 10000），防止 OOM
- `{trace_id, start_time}` 复合索引已建，确认执行计划命中索引

---

## 里程碑建议

| 阶段 | 内容 | 依赖 |
|---|---|---|
| M1 | ✅ #1 Trace 搜索列表 | 无 |
| M2 | ✅ #2 树结构（方案 B） | M1 |
| M3 | #3 聚合统计 | 无 |
| M4 | #6 OTLP 接入 | 无 |
| M5 | #7 采样限流 + #10 metrics + #9 保留策略 | 无 |
