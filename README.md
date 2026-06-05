# Luminous

高校教务系统通用 API 网关，为**光序 Lumaris**提供多校统一注册与发现接口。

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 配置 | 环境变量 / `.env` 文件 |
| 日志 | `log/slog` |


## 目录结构

```
Luminous/
├── cmd/
│   ├── server/main.go            # HTTP 服务入口
│   └── migrate/main.go           # JSON → PostgreSQL 数据迁移工具
├── data/
│   └── schools.json              # 种子数据（4 所西安高校）
├── internal/
│   ├── config/
│   │   └── config.go             # Viper 配置加载与结构体定义
│   ├── handler/
│   │   ├── school.go             # 公开学校接口（列表 / 详情）
│   │   ├── admin.go              # 管理员 CRUD 接口（需认证）
│   │   ├── app.go                # App 版本信息代理接口
│   │   ├── school_service.go     # 学校服务接口定义（扩展预留）
│   │   └── school_test.go        # 公开接口单元测试
│   ├── middleware/
│   │   ├── auth.go               # Bearer Token 认证中间件
│   │   ├── cors.go               # CORS 跨域中间件
│   │   ├── ratelimit.go          # 令牌桶速率限制中间件
│   │   └── request_id.go         # X-Request-ID 生成 / 转发中间件
│   ├── model/
│   │   ├── school.go             # School 实体、Feature 枚举、请求结构体
│   │   └── releaseInfo.go        # App 版本发布信息结构体
│   ├── repository/
│   │   ├── school_repo.go        # SchoolRepository 接口 + JSON 文件实现
│   │   ├── school_repo_pg.go     # PostgreSQL 实现（pgxpool）
│   │   ├── school_repo_test.go   # JSON 仓库单元测试
│   │   └── school_repo_pg_test.go# PG 仓库集成测试（build tag: integration）
│   ├── response/
│   │   └── response.go           # 统一 JSON 响应格式（Success / Error / SuccessList）
│   ├── router/
│   │   └── router.go             # Gin 路由注册与中间件挂载
│   └── util/
│       └── httpclient.go         # HTTP 客户端（重试、User-Agent 轮换、超时）
├── config.yaml                   # 实际配置（Git ignored）
├── Dockerfile                    # 多阶段 Docker 构建
├── Makefile                      # run / build / clean / test 目标
├── go.mod
├── go.sum
└── LICENSE                       # GPLv3
```


## 各文件详细说明

### 入口层 (`cmd/`)

| 文件 | 说明 |
|------|------|
| `cmd/server/main.go` | 服务主入口。按顺序：加载配置 → 设置 Gin 模式 → 连接 PostgreSQL 并自动建表 → 创建三层 handler → 注册路由 → 启动 HTTP 监听 → 监听 SIGINT/SIGTERM 实现优雅关闭（10s 超时）。 |
| `cmd/migrate/main.go` | 一次性数据迁移工具。读取 `data/schools.json` 中的种子数据，逐条 INSERT 到 PostgreSQL（`ON CONFLICT DO NOTHING` 跳过重复）。运行结束后打印成功 / 失败计数。 |

### 配置层 (`internal/config/`)

| 文件 | 说明 |
|------|------|
| `config/config.go` | 定义 `AppConfig` 全局配置结构体（含 `ServerConfig`、`AuthConfig`、`DatabaseConfig`、`ReleaseConfig`）。`LoadConfig()` 使用 Viper 从 `config.yaml` 加载，设置环境变量前缀 `LUMINOUS_`，为所有字段提供合理默认值。 |

### 模型层 (`internal/model/`)

| 文件 | 说明 |
|------|------|
| `model/school.go` | 核心数据模型。`School` 结构体（code、name、website、features、enabled、时间戳）；`Feature` 字符串枚举（10 种教务功能）及 `IsValidFeature()` 校验；`CreateSchoolRequest`（全部必填）和 `UpdateSchoolRequest`（指针字段实现部分更新）。 |
| `model/releaseInfo.go` | App 版本信息结构体：`ReleaseInfo`、`AuthorInfo`、`AssetInfo`、`RawApiResponse`。映射上游 App 更新 API 的 JSON 结构。 |

### 仓库层 (`internal/repository/`)

| 文件 | 说明 |
|------|------|
| `repository/school_repo.go` | 定义 `SchoolRepository` 接口（6 个方法，均接受 `context.Context`）。提供 `JSONSchoolRepository` 实现：内存 map + 磁盘 JSON 持久化，使用 `sync.RWMutex` 保证并发安全。适合测试和开发环境。 |
| `repository/school_repo_pg.go` | 生产级 PostgreSQL 实现。`NewPGSchoolRepository()` 解析 DSN（或分字段拼接）、配置 pgxpool 连接池、Ping 验证、自动建表 + 列补全。所有 CRUD 方法使用传入的 context 以支持请求级取消。`ON CONFLICT DO NOTHING` 处理重复插入。 |
| `repository/school_repo_test.go` | JSON 仓库的完整单测覆盖：Create、FindAll、FindByCode、FindEnabled、Update、Delete、CreateDuplicate、PersistenceAcrossInstances。临时文件自动清理。 |
| `repository/school_repo_pg_test.go` | PG 仓库集成测试（`//go:build integration`）。需要真实数据库（默认 `postgres://luminous:luminous@localhost:5432/luminous`），每用例前后自动建表 / 删表。数据库不可用时 `t.Skipf` 优雅跳过。 |

### 处理器层 (`internal/handler/`)

| 文件 | 说明 |
|------|------|
| `handler/school.go` | `SchoolHandler` — 公开查询接口。`ListSchools` 调用 `Repo.FindEnabled()` 返回所有启用学校；`GetSchool` 按 `:code` 路径参数查询单校详情，未找到返回 404。 |
| `handler/admin.go` | `AdminHandler` — 管理后台接口。`AdminListSchools` 调用 `Repo.FindAll()` 返回全部学校，支持 `?page=` 和 `?page_size=` 分页参数（默认 1/50，最大 200）。`CreateSchool` 校验 Feature 有效性后入库（重复返回 409）。`UpdateSchool` 先查后改，支持部分字段更新。`DeleteSchool` 按 code 删除。全部使用 `c.Request.Context()` 传递请求上下文。 |
| `handler/app.go` | `AppHandler` — App 版本代理。`GetTagModel` 向配置的 `release.api_url`（默认拼接 `release.app_uuid` + `release.channel_id`）发起 GET 请求，将上游 JSON 转换为 `[]ReleaseInfo` 标准格式返回。上游不可达时返回 502。 |
| `handler/school_service.go` | `SchoolServiceHandler` 接口定义：`Code() string` + `RegisterRoutes(*gin.RouterGroup)`。为新学校类型接入预留的扩展点——每种学校类型可实现此接口并注册特定路由。 |
| `handler/school_test.go` | 公开接口的单元测试：空列表返回 200、不存在的学校返回 404、Feature 枚举校验。使用 Gin 测试模式和 JSON 仓库避免外部依赖。 |

### 中间件层 (`internal/middleware/`)

| 文件 | 说明 |
|------|------|
| `middleware/auth.go` | Bearer Token 鉴权。检查 `config.Cfg.Auth.AdminToken` 是否配置（未配置返回 503），解析 `Authorization: Bearer <token>` 头，使用 `crypto/subtle.ConstantTimeCompare` 常量时间比较防时序攻击。 |
| `middleware/cors.go` | CORS 跨域处理。`Access-Control-Allow-Origin` 从 `server.cors_origin` 配置读取（默认 `*`）。允许 GET/POST/PUT/DELETE/OPTIONS 方法，Content-Type 和 Authorization 头。OPTIONS 预检请求直接返回 204，缓存 86400s。 |
| `middleware/ratelimit.go` | 令牌桶速率限制。每 IP 独立计数，默认速率 10 req/s、突发容量 30。桶以秒为单位自动填充，后台协程每 5 分钟清理过期访客记录。超限返回 429。 |
| `middleware/request_id.go` | 请求链路追踪。优先从 `X-Request-ID` 请求头读取，无则用 `crypto/rand` 生成 8 字节随机数（16 位 hex）。`rand.Read` 极端失败时回退为纳秒时间戳，避免重复。写入响应头并存入 Gin Context。 |

### 辅助层

| 文件 | 说明 |
|------|------|
| `response/response.go` | 统一 JSON 响应格式。`Success()` 返回 `{"code":N, "message":"...", "data":...}`；`Error()` 的 data 为 null；`SuccessList()` 包装 `{"total":N, "items":[...]}` 结构。 |
| `router/router.go` | 路由注册中心。组装中间件链（Logger → Recovery → RequestID → CORS → RateLimit）→ 注册 `/healthz` 健康检查 → 挂载公开路由（`/api/v1/schools`、`/api/v1/schools/:code`、`/api/v1/App`）→ 管理员路由组（`/api/v1/admin/*`，额外应用 AuthMiddleware）。 |
| `util/httpclient.go` | HTTP 客户端封装。30s 超时，最多 3 次重试（仅对超时 / `DeadlineExceeded` / 5xx 重试，带指数退避 1s→2s→4s）。每次请求随机选取 10 个常见 User-Agent 之一。暴露全局 `DefaultClient` 单例。支持 GET（含 Cookie / 自定义头）和 POST Form（支持重试的 `GetBody`）。 |

### 根文件

| 文件 | 说明 |
|------|------|
| 环境变量 | 全部配置通过 `LUMINOUS_*` 环境变量注入，详见配置参考章节。本地开发可创建 `config.yaml` 并通过 `LUMINOUS_CONFIG_PATH=./config.yaml` 加载。 |
| `Dockerfile` | 多阶段构建。阶段一用 `golang:1.26-alpine` 编译（`CGO_ENABLED=0`、strip）。阶段二用 `alpine:latest`，安装 ca-certificates + tzdata，复制二进制文件。暴露 8080 端口。 |
| `Makefile` | 四个目标：`run`（`go run ./cmd/server/`）、`build`（输出 `bin/luminous`）、`clean`（删除 `bin/`）、`test`（`go test ./...`）。 |
| `data/schools.json` | 4 所西安高校种子数据：NWPU（西北工业大学）、XAUAT（西安建筑科技大学）、XDU（西安电子科技大学）、XJTU（西安交通大学），含各自的 features 配置。 |


## 架构

```
请求 → Logger → Recovery → RequestID → CORS → RateLimit
                                                      │
                                    ┌─────────────────┼──────────────────┐
                                    ▼                  ▼                  ▼
                              /api/v1/*        /api/v1/admin/*        /healthz
                              (公开路由)         (AuthMiddleware)
                                    │                  │
                                    ▼                  ▼
                              SchoolHandler      AdminHandler
                              AppHandler
                                    │                  │
                                    └────────┬─────────┘
                                             ▼
                                    SchoolRepository (interface)
                                    ┌─────────┴──────────┐
                                    ▼                    ▼
                          JSONSchoolRepository    PGSchoolRepository
                              (内存+文件)          (pgxpool → PostgreSQL)
```


## 快速开始

### 前置条件

- Go 1.26+
- PostgreSQL（本地或远程）
- 可选：Make（Linux/macOS），Windows 下直接用 `go` 命令

### 1. 克隆项目

```bash
git clone <repo-url>
cd Luminous
```

### 2. 配置数据库

确保 PostgreSQL 运行中，创建数据库：

```bash
createdb luminous
# 或通过 psql:
# CREATE DATABASE luminous;
```

### 3. 本地调试

#### 方式一：config.yaml（推荐）

在项目根目录创建 `config.yaml`（已在 `.gitignore` 中忽略，不会提交）：

```yaml
server:
  mode: debug          # debug 模式会输出详细请求日志
  cors_origin: "*"

auth:
  admin_token: "my-dev-token"

database:
  dsn: "postgresql://postgres:postgres@localhost:5432/luminous?sslmode=disable"
  pool_max_conns: 10
  pool_min_conns: 2
```

设置环境变量指向该文件，然后启动：

```bash
# Linux / macOS
export LUMINOUS_CONFIG_PATH="./config.yaml"
go run ./cmd/server/

# Windows PowerShell
$env:LUMINOUS_CONFIG_PATH = ".\config.yaml"
go run ./cmd/server/
```

启动后输出：

```
INFO Starting Luminous server
INFO Server listening addr=:8080
WARN TLS not configured — use a reverse proxy for production
```

#### 方式二：纯环境变量

不创建文件，直接通过环境变量配置：

```bash
# Linux / macOS
export LUMINOUS_SERVER_MODE=debug
export LUMINOUS_DATABASE_DSN="postgresql://postgres:postgres@localhost:5432/luminous?sslmode=disable"
export LUMINOUS_AUTH_ADMIN_TOKEN="my-dev-token"
go run ./cmd/server/

# Windows PowerShell
$env:LUMINOUS_SERVER_MODE = "debug"
$env:LUMINOUS_DATABASE_DSN = "postgresql://postgres:postgres@localhost:5432/luminous?sslmode=disable"
$env:LUMINOUS_AUTH_ADMIN_TOKEN = "my-dev-token"
go run ./cmd/server/
```

### 4. 导入种子数据

```bash
go run ./cmd/migrate/ ./data/schools.json
# 输出: migrated=4 failed=0
```

### 5. 验证服务

```bash
# 健康检查
curl http://localhost:8080/healthz
# → {"status":"ok"}

# 查询学校列表
curl http://localhost:8080/api/v1/schools

# 查询单个学校
curl http://localhost:8080/api/v1/schools/XAUAT

# 管理员接口（需认证）
curl http://localhost:8080/api/v1/admin/schools \
  -H "Authorization: Bearer my-dev-token"
```

默认端口 `8080`，通过环境变量或 `.env` 文件配置。

```bash
# 单元测试（无需数据库）
go test ./... -short

# 集成测试（需 config.yaml 或 LUMINOUS_DATABASE_DSN）
go test -tags=integration ./internal/repository/

# 全部检查
go vet ./...
```

### 生产环境

```bash
# 必需配置
export LUMINOUS_DATABASE_DSN="postgresql://user:password@host:port/dbname?sslmode=require"
export LUMINOUS_AUTH_ADMIN_TOKEN="your-admin-secret-token"

# 启动服务（首次启动自动建表）
./luminous

# 验证
curl http://localhost:8080/healthz
```

## 配置参考

所有配置通过 `LUMINOUS_` 前缀环境变量注入（`.` 替换为 `_`）。YAML 文件仅作本地开发辅助，通过 `LUMINOUS_CONFIG_PATH` 可选加载。

### 服务配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `LUMINOUS_SERVER_PORT` | `8080` | HTTP 监听端口 |
| `LUMINOUS_SERVER_MODE` | `release` | Gin 运行模式（debug / release / test） |
| `LUMINOUS_SERVER_CORS_ORIGIN` | `""` | CORS 允许来源（生产环境应设为具体域名） |
| `LUMINOUS_SERVER_TLS_CERT` | `""` | TLS 证书路径（与 TLS_KEY 同时设置启用 HTTPS） |
| `LUMINOUS_SERVER_TLS_KEY` | `""` | TLS 私钥路径 |
| `LUMINOUS_SERVER_TRUSTED_PROXIES` | `""` | 反向代理 CIDR，逗号分隔（如 `10.0.0.0/8,172.16.0.0/12`）。未设置时 `ClientIP` 取直连 IP |

### 认证配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `LUMINOUS_AUTH_ADMIN_TOKEN` | `""` | 管理员 Bearer Token（未配置时 admin 路由返回 503） |

### 发布信息配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `LUMINOUS_RELEASE_API_URL` | `""` | 上游 App 信息 API 完整 URL（设置后覆盖以下两项） |
| `LUMINOUS_RELEASE_APP_UUID` | `5f278ffc-...` | App UUID |
| `LUMINOUS_RELEASE_CHANNEL_ID` | `9e1a198a-...` | 渠道 ID |

### 数据库配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `LUMINOUS_DATABASE_DSN` | `""` | PostgreSQL 连接串（必填） |
| `LUMINOUS_DATABASE_POOL_MAX_CONNS` | `20` | 连接池最大连接数 |
| `LUMINOUS_DATABASE_POOL_MIN_CONNS` | `5` | 连接池最小连接数 |

### 速率限制配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `LUMINOUS_RATE_LIMIT_RATE` | `10` | 令牌桶填充速率（令牌/秒） |
| `LUMINOUS_RATE_LIMIT_BURST` | `30` | 令牌桶最大容量（突发请求数） |


## API 文档

**Base URL:** `http://localhost:8080`
**Content-Type:** `application/json`

### 通用响应格式

```json
// 单条数据
{ "code": 200, "message": "success", "data": {...} }

// 错误
{ "code": 404, "message": "school not found", "data": null }

// 列表
{ "code": 200, "message": "success", "data": { "total": 3, "items": [...] } }
```

### 公开接口

#### `GET /healthz`

健康检查。

```bash
curl http://localhost:8080/healthz
# → {"status":"ok"}
```

#### `GET /api/v1/schools`

返回所有已启用学校列表。

```bash
curl http://localhost:8080/api/v1/schools
```

#### `GET /api/v1/schools/:code`

返回指定学校详情及支持的功能。

```bash
curl http://localhost:8080/api/v1/schools/XAUAT
```

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 404 | 学校不存在 |

#### `GET /api/v1/App`

获取 App 版本更新信息（代理上游 API）。

```bash
curl http://localhost:8080/api/v1/App
```

### 管理员接口

所有接口挂载在 `/api/v1/admin/` 下，需 Bearer Token 鉴权：

```
Authorization: Bearer <admin_token>
```

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/schools` | 列出所有学校（含未启用），支持分页 |
| POST | `/api/v1/admin/schools` | 新增学校 |
| PUT | `/api/v1/admin/schools/:code` | 部分更新学校 |
| DELETE | `/api/v1/admin/schools/:code` | 删除学校 |

**分页参数：** `GET /api/v1/admin/schools?page=1&page_size=50`（page 默认 1，page_size 默认 50，最大 200）

复制 `.env.example` 为 `.env` 并修改：

```bash
cp .env.example .env
```

```ini
LUMINOUS_SERVER_PORT=8080
LUMINOUS_SERVER_MODE=debug
LUMINOUS_AUTH_ADMIN_TOKEN=your-admin-secret-token
LUMINOUS_DATA_SCHOOLS_FILE=./data/schools.json
```

环境变量优先级高于 `.env` 文件。`LUMINOUS_AUTH_ADMIN_TOKEN` **必须设置**，否则服务拒绝启动。


```text
Luminous/
├── cmd/server/main.go
├── .env.example
├── data/schools.json
├── internal/
│   ├── config/config.go
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── response/
│   ├── router/
│   └── util/
├── go.mod
└── README.md
```
