# Mallard — 链路追踪系统

基于 Echo v5 + MongoDB 的调用链追踪系统：业务应用通过 App 凭证批量上报 span，按 trace_id 检索链路、搜索 Trace、聚合分析调用关系。

## Web 控制台（Mallard-UI）

管理后台（登录 / Trace 检索 / Trace 详情 / App 管理）由独立前端项目 [mallard-ui](https://github.com/li-qs/mallard-ui)（React 18 + TypeScript + Vite + Ant Design 5）提供。

- **构建产物已内嵌进 server 二进制**（`//go:embed internal/web/dist`），启动 `cmd/mallard` 后直接访问 `http://localhost:9010/` 即可使用，无需单独起前端
- 更新 UI：在 mallard-ui 仓库 `npm run build` 后把 `dist/` 拷贝到本仓库 `internal/web/dist/` 再重新编译
- 本地联调：前端 dev server 默认 `http://localhost:5173`；后端需在 `config.yaml` 的 `allow_origins` 加上该源，并设 `cookie_secure: false`
- 对接说明：本仓库 [mallard-ui-design.md](./mallard-ui-design.md)（API 契约、页面设计、全局约定）

## 技术栈

- **框架**: [Echo v5](https://github.com/labstack/echo)
- **数据库**: [MongoDB](https://www.mongodb.com/) ([mongo-driver v2](https://github.com/mongodb/mongo-go-driver))
- **日志**: [Zap](https://github.com/uber-go/zap)
- **校验**: [validator/v10](https://github.com/go-playground/validator)
- **配置**: [Viper](https://github.com/spf13/viper)（支持环境变量覆盖，如 `AUTH_JWT_SECRET`）
- **认证**: [JWT](https://github.com/golang-jwt/jwt) (Access Token) + Refresh Token，bcrypt 密码哈希

## 项目结构

```
├── cmd/server/           # 入口（装配根：组装依赖 + 启动）
├── internal/
│   ├── config/           # 配置加载、默认值、校验
│   ├── dto/              # 请求体
│   ├── handler/          # HTTP 处理
│   ├── logger/           # Zap 日志封装
│   ├── middleware/       # 认证、请求日志、限流
│   ├── model/            # MongoDB 模型（含 DB/Collection 常量）
│   ├── repository/       # 数据访问层（构造函数无副作用，索引迁移收口在 Migrate）
│   ├── response/         # 泛型 JSON 响应封装
│   ├── router/           # 路由注册
│   ├── service/          # 业务逻辑
│   ├── utils/            # 工具函数
│   └── vo/               # 响应体
```

> 依赖通过 `cmd/server/main.go` 组装（composition root）：构造 repository → `repository.Migrate` 统一建索引 → 组装 service → 以 `router.Deps` 注入路由层。
> 数据分两个库：`mallard`（user / refresh_token）、`mallard_tracing`（app / span），启动时 `Migrate` 自动建索引。

## 认证流程

```
登录:   POST /api/login     → 返回 JWT（响应体） + Refresh Token（HttpOnly Cookie）
刷新:   POST /api/refresh   → Cookie 携带 Refresh Token，返回新 JWT + 新 Refresh Token（Rotation）
登出:   POST /api/logout    → 删除 Refresh Token，清除 Cookie
改密:   PUT /api/user/password → 撤销该用户所有 Refresh Token，需重新登录
```

- Access Token（JWT）有效期 15 分钟，放 `Authorization: Bearer <token>` Header
- Refresh Token 为随机不透明串，有效期 7 天，服务端只存 SHA-256 哈希
- `/api/login` 按 IP 限流（默认 10 req/s），`cookie_secure` 默认开启，纯 HTTP 本地联调时需显式设为 `false`

### App 上报流程

```
注册 App:  POST /api/app          → 返回 app_id + secret（仅展示一次）
配置 allow list: PUT /api/app/:id/ip-allow-list → 限制上报来源 IP（CIDR 或精确 IP）
上报 Span:  POST /api/v1/spans    → collector :9011，Authorization: Basic base64(app_id:secret)，批量上报
检索链路:  GET /api/traces        → server :9010，按条件搜索 Trace；GET /api/traces/:trace_id 查详情
```

- App secret 为 32 字节随机串，服务端只存 SHA-256 哈希，本地 TTL 缓存（60s）+ `subtle.ConstantTimeCompare` 校验
- `app_id` 由服务端从凭证解析，不信任客户端传参；allow list 为空则不限制 IP
- 上报采用批量接口（`span_id` 唯一索引幂等去重，重复上报自动忽略）

## 快速开始

```bash
# 1. 修改 config.yaml 中的 mongodb.main 连接串和 jwt_secret

# 2. 编译运行后端（启动时会自动建索引并校验数据库连通性）
go run ./cmd/server

# 或指定配置文件
go run ./cmd/server -config /path/to/config.yaml

# 3. 启动 Web 控制台（独立项目，见上文「Web 控制台」）
git clone https://github.com/li-qs/mallard-ui
cd mallard-ui && npm install && npm run dev   # 浏览器访问 http://localhost:5173
```

## 编译

```bash
# 交叉编译 linux / macos (amd64, arm64)
make

# 或指定目标
make server-linux-amd64
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
| PUT | `/api/user/password` | 修改密码 | 是 |
| POST | `/api/app` | 新增 App，返回 app_id + secret（仅展示一次） | 是 |
| GET | `/api/app` | App 列表（`app_name` 模糊 / `id` 精确，可组合；分页：`page`、`page_size`） | 是 |
| PUT | `/api/app/:id/ip-allow-list` | 更新 App IP allow list | 是 |
| PUT | `/api/app/:id/secret` | 轮换 App secret | 是 |
| DELETE | `/api/app/:id` | 删除 App | 是 |
| POST | `/api/v1/spans` | 批量上报 Span（collector，`Authorization: Basic base64(app_id:secret)`） | App |
| GET | `/api/traces/:trace_id` | 查询某个 trace 的全部 span，按 start_time 正序（含 `is_root` 根标记） | 是 |
| GET | `/api/traces` | Trace 搜索列表（`app_id`/`operation`/`status`(1=成功 2=错误)/`trace_id`/`start_time_gt`/`start_time_lt`/`page`/`page_size`） | 是 |

> App 管理、Trace 查询接口走用户 JWT（`Authorization: Bearer <token>`）；span 上报走 App 凭证（Basic base64 认证）。
> Web 控制台（mallard-ui 构建产物）内嵌在 server 二进制中，访问 `http://localhost:9010/` 直接使用；API 统一 `/api` 前缀。

## 配置

```yaml
server_addr: :9010
collector_addr: :9011

allow_origins:
  - http://localhost:3000
collector_allow_origins:
  - http://localhost:3000

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
