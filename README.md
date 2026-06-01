# Luminous

高校教务系统通用 API 网关，为移动应用提供统一的多校教务数据接口。

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 配置 | 环境变量 / `.env` 文件 |
| 日志 | `log/slog` |
| 缓存 | 内存 TTL Cache (`sync.Map` + 过期时间) |

## 快速开始

```bash
go mod tidy
go run ./cmd/server/
curl http://localhost:8080/api/v1/schools
```

默认端口 `8080`，通过环境变量或 `.env` 文件配置。

## API 概览

### 通用说明

- Base URL: `http://localhost:8080`
- Content-Type: `application/json`
- 统一响应格式:

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

列表接口使用 `data.items` 和 `data.total`：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 3,
    "items": []
  }
}
```

### 学校信息接口

#### `GET /api/v1/schools`

返回所有已启用学校。

```bash
curl http://localhost:8080/api/v1/schools
```

#### `GET /api/v1/schools/:code`

返回指定学校详情。

```bash
curl http://localhost:8080/api/v1/schools/NWPU
```

常见状态码：

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 404 | 学校不存在 |

### App 接口

#### `GET /api/v1/App`

代理 Gitee Release 接口，返回原始 JSON。

#### `GET /api/v1/App/GetTag`

返回转换后的版本信息模型。

### 管理员接口

所有接口挂载在 `/api/v1/admin/` 下，需要：

```text
Authorization: Bearer <admin_token>
```

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/schools` | 列出所有学校（含未启用） |
| POST | `/api/v1/admin/schools` | 新增学校 |
| PUT | `/api/v1/admin/schools/:code` | 更新学校（部分更新） |
| DELETE | `/api/v1/admin/schools/:code` | 删除学校 |

## 配置

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

## 项目结构

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
