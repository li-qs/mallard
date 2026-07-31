# go-web-framework

基于 Echo v5 的 Web API 框架，用于快速搭建新项目。

## 技术栈

- **框架**: [Echo v5](https://github.com/labstack/echo)
- **数据库**: MySQL ([sqlx](https://github.com/jmoiron/sqlx))
- **日志**: [Zap](https://github.com/uber-go/zap)
- **校验**: [validator/v10](https://github.com/go-playground/validator)
- **认证**: [JWT](https://github.com/golang-jwt/jwt) (Access Token) + Refresh Token，bcrypt 密码哈希

## 项目结构

```
├── cmd/server/           # 入口
├── internal/
│   ├── config/           # 配置加载、默认值、校验
│   ├── dto/              # 请求体
│   ├── handler/          # HTTP 处理
│   ├── logger/           # Zap 日志封装
│   ├── middleware/       # 认证中间件
│   ├── model/            # 数据库模型
│   ├── repository/       # 数据访问层
│   ├── response/         # 泛型 JSON 响应封装
│   ├── router/           # 路由注册 & 依赖注入
│   ├── service/          # 业务逻辑
│   ├── utils/            # 工具函数
│   └── vo/               # 响应体
├── migration/            # SQL 迁移文件
└── pkg/mysql/            # MySQL 连接封装
```

## 认证流程

```
登录:   POST /login     → 返回 JWT（响应体） + Refresh Token（HttpOnly Cookie）
刷新:   POST /refresh   → Cookie 携带 Refresh Token，返回新 JWT + 新 Refresh Token（Rotation）
登出:   POST /logout    → 删除 Refresh Token，清除 Cookie
改密:   PUT /user/password → 撤销该用户所有 Refresh Token，需重新登录
```

- Access Token（JWT）有效期 15 分钟，放 `Authorization: Bearer <token>` Header
- Refresh Token 为随机不透明串，有效期 7 天，服务端只存 SHA-256 哈希

## 快速开始

```bash
# 1. 创建数据库并执行迁移
mysql -u root -e "CREATE DATABASE myapi CHARACTER SET utf8mb4"
mysql -u root myapi < migration/001_user.sql
mysql -u root myapi < migration/002_refresh_token.sql

# 2. 修改 config.yaml 中的数据库连接和 jwt_secret

# 3. 编译运行
go run ./cmd/server

# 或指定配置文件
go run ./cmd/server /path/to/config.yaml
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
| POST | `/login` | 登录 | 否 |
| POST | `/logout` | 登出 | 否 |
| POST | `/refresh` | 刷新 Token | 否 |
| GET | `/health` | 健康检查 | 否 |
| GET | `/user` | 当前用户信息 | 是 |
| PUT | `/user/password` | 修改密码 | 是 |

## 配置

```yaml
server:
  listen_addr: :8080
  allow_origins:
    - http://localhost:3000

database:
  main: user:password@tcp(localhost:3306)/myapi?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai&timeout=5s

auth:
  jwt_secret: "your-random-secret"
  access_token_expire_seconds: 900
  refresh_token_expire_seconds: 604800
```

- `listen_addr` 默认 `:8080`
- `access_token_expire_seconds` 默认 `900`（15 分钟）
- `refresh_token_expire_seconds` 默认 `604800`（7 天）
