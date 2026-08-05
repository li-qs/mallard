# mallard-ui 前台页面设计书

> 面向 AI 前端开发：本设计书用于搭建 mallard-ui（React 前台），并与 mallard 后端 API 对接。
> 后端 API 以 README.md 与本文「API 契约」章节为准；字段名、单位、错误处理约定必须严格遵循。

---

## 1. 项目概述

**定位**：链路追踪系统（tracing）的 Web 控制台。业务应用通过 App 凭证上报 span，运营人员在此检索 Trace、查看链路详情、管理上报 App。

**核心页面**（5 个）：
1. 登录页 `/login`
2. Trace 列表页 `/traces`（搜索 + 分页）
3. Trace 详情页 `/traces/:traceId`（span 列表 / 瀑布图）
4. App 管理页 `/apps`
5. 个人中心 `/account`（当前用户信息 + 修改密码）

**布局**：登录页独立无侧栏；其余页面共用侧边栏布局（顶栏含用户名/退出）。

**建议技术栈**：
| 项 | 选型 | 说明 |
|---|---|---|
| 框架 | React 18 + TypeScript | |
| 构建 | Vite | |
| UI 组件库 | Ant Design 5 | 表格/表单/弹窗直接用 AntD |
| 路由 | react-router-dom v6 | 受保护路由需登录 |
| HTTP | axios | 统一拦截器 |
| 状态 | zustand（或 React Context） | 存登录态/用户信息 |
| 时间 | dayjs | 时间戳格式化 |

---

## 2. 项目目录结构

```
src/
├── api/               # 接口层（每个模块一个文件）
│   ├── http.ts        # axios 实例 + 拦截器
│   ├── auth.ts        # 登录/登出/刷新/用户信息/改密
│   ├── app.ts         # App 管理
│   └── trace.ts       # Trace 搜索/详情
├── types/             # 全局 TS 类型（与后端契约一一对应）
│   └── index.ts
├── components/        # 通用组件
│   ├── ProtectedRoute.tsx
│   └── SpanTree.tsx   # Trace 详情瀑布图/树
├── pages/
│   ├── Login.tsx
│   ├── TraceList.tsx
│   ├── TraceDetail.tsx
│   ├── AppList.tsx
│   └── Account.tsx
├── layouts/
│   └── MainLayout.tsx # 侧边栏 + 顶栏 + <Outlet/>
├── store/
│   └── auth.ts        # zustand：token、user、login/logout
├── router/
│   └── index.tsx      # 路由表 + ProtectedRoute
└── utils/
    └── format.ts      # 时间/纳秒→毫秒/延迟格式化
```

---

## 3. 全局约定（AI 必须严格遵守）

### 3.1 响应格式

所有接口返回 **HTTP 200 + JSON body**，body 统一结构：

```ts
// 普通响应
interface ApiResponse<T> { code: number; message?: string; data: T }
// 分页响应
interface ListData<T> { page: number; page_size: number; total: number; list: T[] }
```

- `code === 0` 表示成功；**非 0 视为业务失败**，弹 `message`。
- 例外：`/ready` 返回 HTTP 503；登录限流触发时返回 HTTP 429；校验失败返回 HTTP 400。
- **axios 响应拦截器规则**：
  1. `res.status` 非 2xx（401/403/429/5xx）→ 按错误处理（401 触发重新登录流程）
  2. HTTP 200 时判断 `body.code !== 0` → 按业务错误处理，`message` 提示
  3. 通过后返回 `body.data`

### 3.2 认证流程（重要）

- 登录成功：`access_token` 放在响应体 `data.access_token`，`refresh_token` 通过 **HttpOnly Cookie** 自动设置（前端不可读、无需处理）。
- **每次请求**需在 `Authorization: Bearer <access_token>` 头携带 access_token。
- 会话保持策略：受保护请求返回 401 时，先调 `POST /refresh`（Cookie 自动带上）换取新 access_token 后重试原请求；refresh 也失败则清登录态跳 `/login`。
- 退出登录：调 `POST /logout`（清服务端 refresh token + Cookie），并清前端 token。

### 3.3 时间单位（极易出错）

| 字段 | 单位 |
|---|---|
| `created_at` / `updated_at` / `reported_at` | **毫秒**（int64） |
| `start_time` / `duration`（span 与 trace_summary 均有） | **纳秒**（int64） |

前端统一封装 `utils/format.ts`：
- `msToTime(ms)`：毫秒时间戳 → `YYYY-MM-DD HH:mm:ss`（dayjs）
- `nsToMs(ns)`：`ns / 1e6`
- `formatDuration(ns)`：纳秒 → 可读时长（`3.5ms` / `1.2s`），小于 1ms 用 `µs`
- 瀑布图偏移量统一用纳秒计算，展示时转毫秒

### 3.4 状态与文案约定

- span/trace 的错误定义：`status !== 0` 视为错误（`has_error` / `error_count`）。列表页用红色 Tag 标记错误 Trace。
- 空状态：AntD `Table` 的 `locale.emptyText` 显示「暂无数据」。
- 全局提示用 AntD `message`；二次确认用 `Modal.confirm`。
- 分页参数名：`page`（从 1 起）、`page_size`（默认 10，后端上限 100）。

---

## 4. API 契约（TypeScript 类型 + 接口清单）

### 4.1 公共类型（src/types/index.ts）

```ts
export interface ApiResponse<T> { code: number; message?: string; data: T }
export interface ListData<T> { page: number; page_size: number; total: number; list: T[] }

// ---- auth ----
export interface LoginReq { username: string; password: string }
export interface LoginRes { access_token: string; token_type: string; expires_in: number }
export interface UpdatePasswordReq { password: string; new_password: string }

// ---- user ----
export interface User { id: string; username: string; created_at: number; updated_at: number }

// ---- app ----
export interface AddAppReq { app_name: string; ip_allow_list?: string[] }
export interface App { id: string; app_name: string; ip_allow_list: string[]; created_at: number; updated_at: number }
export interface AppAddRes { id: string; app_name: string; secret: string; ip_allow_list: string[] }
export interface AppSecretRes { id: string; secret: string }

// ---- span ----
export interface Span {
  id: string; app_id: string; trace_id: string; span_id: string;
  parent_id: string; is_root: boolean; operation: string;
  start_time: number; duration: number; status: number; error?: string; reported_at: number;
}
export interface TraceSummary {
  trace_id: string; app_ids: string[]; operation: string;
  start_time: number; duration: number; span_count: number; error_count: number; has_error: boolean;
}
```

### 4.2 接口清单

| 方法 | 路径 | 认证 | 说明 | 入参 | 出参(data) |
|---|---|---|---|---|---|
| POST | `/api/login` | 否 | 登录（Cookie 自动带 refresh_token） | `LoginReq` | `LoginRes` |
| POST | `/api/logout` | 否 | 登出（清 Cookie） | - | `""` |
| POST | `/api/refresh` | 否 | 刷新 access_token（靠 Cookie） | - | `LoginRes` |
| GET | `/api/user` | JWT | 当前用户 | - | `User` |
| PUT | `/api/user/password` | JWT | 修改密码（成功后后端撤销该用户全部 refresh token，需重新登录） | `UpdatePasswordReq` | `""` |
| POST | `/api/app` | JWT | 新增 App，`secret` **仅返回一次** | `AddAppReq` | `AppAddRes` |
| GET | `/api/app` | JWT | App 列表（`app_name` 子串模糊 / `id` 精确，可组合 AND；`page`、`page_size` 分页） | `page, page_size, app_name, id` | `ListData<App>` |
| PUT | `/api/app/:id/ip-allow-list` | JWT | 更新 IP allow list | `{ ip_allow_list: string[] }` | `""` |
| PUT | `/api/app/:id/secret` | JWT | 轮换 secret，新 secret 仅返回一次 | - | `AppSecretRes` |
| DELETE | `/api/app/:id` | JWT | 删除 App | - | `""` |
| GET | `/api/traces/:traceId` | JWT | 某 trace 全部 span（start_time 正序） | - | `Span[]` |
| GET | `/api/traces` | JWT | Trace 搜索列表 | 见 4.3 | `ListData<TraceSummary>` |

> 注：`POST /api/v1/spans` 是 App 上报接口（Basic 认证，走 collector），**前端控制台不使用**，仅了解即可。

### 4.3 Trace 搜索参数（GET /api/traces）

全部为 query 参数，均可选：
```
app_id          string   # App ID（精确）
operation       string   # 操作名（精确）
status          number   # Trace 级筛选：不传=全部，1=成功（无错误 span），2=错误（存在错误 span）
trace_id        string   # trace_id 前缀模糊匹配
start_time_gt   number   # 开始时间下界（纳秒）
start_time_lt   number   # 开始时间上界（纳秒）
page            number   # 默认 1
page_size       number   # 默认 10，上限 100
```

---

## 5. 页面设计

### 5.1 登录页 `/login`

- 布局：居中卡片，标题「mallard 链路追踪」，表单 `用户名` + `密码`。
- 交互：
  - 提交 → `POST /login`；成功后把 `access_token` 存 zustand，跳转 `/traces`。
  - 失败（code 非 0，如「用户名或密码错误」）→ 表单内红色提示；不弹全局 message。
  - 密码错误与用户不存在统一文案，无需区分。
- 边界：登录按钮 loading 防重复提交。

### 5.2 布局 `MainLayout`

- 左侧菜单：**Trace 检索**（`/traces`）、**App 管理**（`/apps`）、**个人中心**（`/account`）。
- 顶栏右侧：用户名 + 下拉「退出登录」（调 `/logout` → 清 token → 跳 `/login`）。
- 路由用 `<Outlet/>` 承载子页面。

### 5.3 Trace 列表页 `/traces`

**筛选表单**（顶部，支持折叠）：
- `trace_id`（Input，placeholder「trace_id 前缀」）
- `app_id`（Input）
- `operation`（Input）
- `status`（Select：全部(不传) / 成功(1) / 错误(2)）
- `start_time`（`RangePicker`，选时间范围后转换为纳秒传 `start_time_gt` / `start_time_lt`）

**表格列**：
| 列 | 取值 | 渲染 |
|---|---|---|
| trace_id | `trace_id` | 链接到 `/traces/:trace_id` |
| 操作 | `operation` | 文本（根 span 的 operation） |
| 所属 App | `app_ids` | 数组 join(' / ') |
| 开始时间 | `start_time` | `nsToMs` 后 `msToTime` |
| 耗时 | `duration` | `formatDuration` |
| span 数 | `span_count` | 数字 |
| 错误 | `has_error` | 是→红色 Tag「错误」；否→灰色「正常」 |

- 分页：AntD `Table` 的 `pagination`（`current=page`、`pageSize=page_size`、`total`），onChange 重新请求。
- 交互：搜索按钮 → 重置到第 1 页再查询；「重置」清空表单并刷新。
- 时间范围换算：RangePicker 值（毫秒）转纳秒 `ms * 1e6` 后再传。

### 5.4 Trace 详情页 `/traces/:traceId`

- 顶部：返回按钮 + trace_id + 概要（span 总数、是否有错误、耗时）。
- 主体：**span 列表（瀑布图 / 树）**，方案 B —— 后端返回平铺列表，前端按 `parent_id` 组树：
  - 数据：`GET /traces/:traceId` 返回 `Span[]`（已按 start_time 正序）。
  - 组树算法：`span_id` → 节点 map；遍历列表，`parent_id === ''`（或 `is_root === true`）为根；其余挂到父节点 `children`。
  - 渲染：`components/SpanTree.tsx`，用 AntD `Table` 展开行或自绘缩进列表：
    - 每行：`operation`、`app_id`、`status`（非 0 红色 Tag）、`formatDuration(duration)`
    - 瀑布条：以整条 trace 的 `min start_time` 为基准，计算每行 `left = (start_time - minStart) / totalDuration` 的偏移，宽度 `duration / totalDuration`，用 CSS 百分比绘制色条
  - 根 span 高亮（`is_root`）。
- 错误 span 展示 `error` 字段内容（Tooltip 或单独列）。

### 5.5 App 管理页 `/apps`

**查询框**（表格上方）：
- `app_name`（Input，子串模糊）、`id`（Input，App ID 精确匹配）
- 搜索/重置/翻页/删除/刷新都**携带当前筛选条件**，搜索后重置到第 1 页

**表格列**：
| 列 | 渲染 |
|---|---|
| app_name | 文本 |
| id | 文本（可复制） |
| ip_allow_list | 空→「不限」；否则 join(' / ') |
| created_at / updated_at | `msToTime` |
| 操作 | 编辑 allow list、轮换 secret、删除 |

**新建 App**（Modal 表单：`app_name` 必填 + `ip_allow_list`（可多行输入，逗号/CIDR））：
- 提交 → `POST /app` → 弹出「**secret 仅展示一次**」结果弹窗，展示 `secret` + 复制按钮，并附 Basic 认证示例：
  `Authorization: Basic base64(app_id:secret)`。关闭后 secret 不可再见，需刷新列表。
- 弹窗需提示「请立即保存，关闭后将无法再次查看」。

**轮换 secret**：行内按钮 → `Modal.confirm`（提示旧 secret 立即失效）→ `PUT /app/:id/secret` → 同样弹「仅展示一次」结果弹窗。

**编辑 allow list**：行内按钮 → Modal 多行输入（一行一条 IP 或 CIDR，如 `192.168.1.10`、`10.0.0.0/24`）→ `PUT /app/:id/ip-allow-list`。

**删除**：行内按钮 → `Modal.confirm`（危险色，确认文案含 app_name）→ `DELETE /app/:id` → 刷新列表。

### 5.6 个人中心 `/account`

- 展示当前用户：`id`、`username`、`created_at`（只读）。
- 修改密码表单：`password`（旧密码）、`new_password`、确认新密码（前端二次确认一致性）。
- 提交成功 → `message.success('密码已修改，请重新登录')` → 清 token 跳 `/login`（后端已撤销全部 refresh token）。

---

## 6. 路由与权限

```
/login            → 公开
/traces           → Protected（MainLayout）
/traces/:traceId  → Protected（MainLayout）
/apps             → Protected（MainLayout）
/account          → Protected（MainLayout）
*                 → 重定向 /traces
```

`ProtectedRoute`：无 token → `<Navigate to="/login" />`；有 token 渲染 `<Outlet/>`。

---

## 7. 联调注意事项

- 后端同源部署：UI 构建产物内嵌进 server 二进制，访问 `http://localhost:9010/` 即用；axios `baseURL` 默认 `/api`（可用 `.env` 的 `VITE_API_BASE_URL` 覆盖，如本地联调指向 `http://localhost:9010/api`）。
- 后端 CORS 已开启 `AllowCredentials`，前端 Cookie 模式需要 `withCredentials: true`（axios）。
- 本地纯 HTTP 联调：后端 `config.yaml` 需设 `cookie_secure: false`（否则 refresh_token Cookie 不会在 http 下发送）。
- 登录态失效（refresh 也失败）统一处理：清 token + 跳登录 + `message.warning('登录已过期')`。

---

## 8. 验收标准

1. 未登录访问受保护页面自动跳转 `/login`。
2. 登录后能进入 Trace 列表，筛选（时间范围/错误）与分页正确，列字段与单位换算无误。
3. Trace 详情页按 `parent_id` 正确组树，根 span 高亮，瀑布条位置/宽度正确，错误 span 展示 `error`。
4. App 新建/轮换 secret 的「仅展示一次」弹窗正确，App 列表查询框（名称模糊/ID 精确）筛选生效，allow list 编辑与删除走二次确认。
5. 修改密码后强制重新登录。
6. token 过期时自动 refresh 并重试，最终失败跳登录。
