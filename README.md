# Mallard — 链路追踪系统

基于 Echo v5 + MongoDB 的调用链追踪系统：业务应用通过 App 凭证批量上报 span，按 trace_id 检索链路、搜索 Trace、聚合分析调用关系。

服务拆分为两个二进制：

- **mallard**（管理端，`:9010`）：登录认证、App 管理、Trace 检索，并内嵌 Web 控制台
- **collector**（采集端，`:9011`）：接收业务应用批量上报 span（App Basic 凭证认证）

## Web 控制台

管理后台（登录 / Trace 检索 / Trace 详情 / App 管理）由独立前端项目 [mallard-ui](https://github.com/li-qs/mallard-ui)（React + TypeScript + Ant Design）提供，构建产物通过 `//go:embed` 内嵌进 mallard 二进制，启动后直接访问 `http://localhost:9010/` 即可使用。

- 更新 UI：在 mallard-ui 仓库 `npm run build` 后把 `dist/` 拷贝到本仓库 `internal/web/dist/` 再重新编译
- 本地联调：前端 dev server 默认 `http://localhost:5173`，需在 `config.yaml` 的 `allow_origins` 加上该源并设 `cookie_secure: false`
- 对接说明：[mallard-ui-design.md](./mallard-ui-design.md)（API 契约、页面设计）；界面预览：[screenshots.md](./screenshots.md)

## 技术栈

- **框架**: [Echo v5](https://github.com/labstack/echo)
- **数据库**: [MongoDB](https://www.mongodb.com/) ([mongo-driver v2](https://github.com/mongodb/mongo-go-driver))
- **认证**: [JWT v5](https://github.com/golang-jwt/jwt)（Access Token）+ Refresh Token Rotation，[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) 密码哈希
- **配置**: [go.yaml.in/yaml/v4](https://go.yaml.in)（`-config` 指定配置文件）
- **校验**: [validator/v10](https://github.com/go-playground/validator)
- **缓存**: [ristretto v2](https://github.com/dgraph-io/ristretto)（App 凭证本地缓存）
- **日志**: 标准库 `log/slog`（JSON 结构化输出）

## 项目结构

```
├── cmd/
│   ├── mallard/             # 管理端入口：装配依赖 + 注册路由 + 内嵌 Web 控制台（:9010）
│   └── collector/           # 采集端入口：装配依赖 + 注册上报路由（:9011）
├── internal/
│   ├── config/              # 配置加载、默认值、校验
│   ├── domain/              # 领域层，按业务域拆分
│   │   ├── health/          # 存活 / 就绪探针
│   │   ├── user/            # 登录认证、用户
│   │   ├── app/             # App 凭证、IP allow list
│   │   └── span/            # span 上报、Trace 检索 / 聚合
│   ├── infra/repo/          # MongoDB 数据访问实现 + Migrate 统一建索引
│   ├── middleware/          # JWT 认证、App Basic 认证、请求日志
│   ├── reqctx/              # 请求上下文（当前用户 / 当前 App）
│   ├── request/             # 分页解析等请求工具
│   ├── response/            # 泛型 JSON 响应封装
│   ├── server/              # Echo server 通用启动（优雅退出）
│   └── web/                 # Web 控制台静态资源内嵌 + SPA 回退
└── pkg/
    ├── mongodb/             # Mongo 客户端连接
    └── utils/               # hash / random 工具
```

## 认证流程

```
登录:   POST /api/login     → 返回 JWT（响应体） + Refresh Token（HttpOnly Cookie）
刷新:   POST /api/refresh   → Cookie 携带 Refresh Token，返回新 JWT + 新 Refresh Token（Rotation）
登出:   POST /api/logout    → 删除 Refresh Token，清除 Cookie
改密:   PUT /api/user/password → 校验旧密码后更新，并撤销该用户全部 Refresh Token
```

- Access Token（JWT）有效期 15 分钟，放 `Authorization: Bearer <token>` Header
- Refresh Token 为随机不透明串，有效期 7 天，服务端只存 HMAC-SHA256（`token_salt` 加盐）哈希
- `/api/login` 按 IP 限流（默认 10 req/s），`cookie_secure` 默认开启，纯 HTTP 本地联调时需显式设为 `false`

### App 上报流程

```
注册 App:  POST /api/app          → 返回 app_id + secret（仅展示一次）
配置 allow list: PUT /api/app/:id/ip-allow-list → 限制上报来源 IP（CIDR 或精确 IP）
上报 Span:  POST /api/v1/spans    → collector :9011，Authorization: Basic base64(app_id:secret)，批量上报
检索链路:  GET /api/traces        → mallard :9010，按条件搜索 Trace；GET /api/traces/:trace_id 查详情
```

- App secret 为 32 字节随机串，服务端只存 SHA-256 哈希，本地 ristretto 缓存 + `subtle.ConstantTimeCompare` 校验
- `app_id` 由服务端从凭证解析，不信任客户端传参；allow list 为空则不限制 IP
- 上报采用批量接口（`span_id` 唯一索引幂等去重，重复上报自动忽略）

## 快速开始

```bash
# 1. 修改 config.yaml 中的 mongo_uri 和 jwt_secret

# 2. 编译运行管理端（:9010，含 Web 控制台，启动自动建索引）
go run ./cmd/mallard

# 3. 编译运行采集端（:9011）
go run ./cmd/collector

# 或指定配置文件
go run ./cmd/mallard -config /path/to/config.yaml
```

Web 控制台已内嵌进 mallard 二进制，启动后直接访问 `http://localhost:9010/` 即可，无需单独启动前端。

## 编译

```bash
# 交叉编译 linux / darwin (amd64, arm64) 的管理端与采集端
make

# 或只构建某个二进制
make server-linux-amd64
make collector-darwin-arm64
```

二进制文件输出到 `build/` 目录。

## API

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 存活探针（liveness），进程存活即 200 | 否 |
| GET | `/ready` | 就绪探针（readiness），Mongo 可达返回 200，否则 503 | 否 |
| POST | `/api/login` | 登录 | 否 |
| POST | `/api/logout` | 登出 | 否 |
| POST | `/api/refresh` | 刷新 Token | 否 |
| GET | `/api/user` | 当前用户信息 | 是 |
| PUT | `/api/user/password` | 修改密码（需校验旧密码） | 是 |
| POST | `/api/app` | 新增 App，返回 app_id + secret（仅展示一次） | 是 |
| GET | `/api/app` | App 列表（`app_name` 模糊 / `id` 精确，可组合；分页：`page`、`page_size`） | 是 |
| PUT | `/api/app/:id/ip-allow-list` | 更新 App IP allow list | 是 |
| PUT | `/api/app/:id/secret` | 轮换 App secret | 是 |
| DELETE | `/api/app/:id` | 删除 App | 是 |
| POST | `/api/v1/spans` | 批量上报 Span（collector，`Authorization: Basic base64(app_id:secret)`） | App |
| GET | `/api/traces/:trace_id` | 查询某个 trace 的全部 span，按 start_time 正序（含 `is_root` 根标记） | 是 |
| GET | `/api/traces` | Trace 搜索列表（`app_id`/`operation`/`status`(1=成功 2=错误)/`trace_id`/`start_time_gt`/`start_time_lt`/`page`/`page_size`） | 是 |

> 管理端接口（:9010）走用户 JWT（`Authorization: Bearer <token>`）；span 上报（:9011）走 App 凭证（Basic base64 认证）。
> 统一 JSON 封装 `{code, message?, data?}`，列表返回 `data: {page, page_size, total, list}`。

## 配置

```yaml
server_addr: :9010
collector_addr: :9011

allow_origins:
  - http://localhost:5173
collector_allow_origins:
  - http://localhost:5173

log_level: info

mongo_uri: mongodb://user:password@localhost:27017/mallard?authSource=admin

jwt_secret: "your-random-secret"
token_salt: "your-random-salt"   # refresh token HMAC 加盐（防彩虹表）
access_ttl: 900
refresh_ttl: 604800
span_ttl: 604800     # span 数据保留时长（秒），通过 TTL 索引自动清理
cookie_secure: true  # 纯 HTTP 本地开发请设为 false
```

- `server_addr` 默认 `:9010`（管理端）
- `collector_addr` 默认 `:9011`（采集端）
- `access_ttl` 默认 `900`（15 分钟，秒）
- `refresh_ttl` 默认 `604800`（7 天，秒）
- `cookie_secure` 默认 `true`
- 必填项：`mongo_uri`、`jwt_secret`；`log_level: debug` 可开启 debug 日志
