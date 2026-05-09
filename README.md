# Luminous

高校教务系统通用 API 网关，为移动应用提供统一的多校教务数据接口。

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 配置 | Viper (YAML + 环境变量) |
| 日志 | `log/slog` |
| 指标 | Prometheus (`/metrics`) |
| 缓存 | 内存 TTL Cache (`sync.Map` + 过期时间) |

## 快速开始

```bash
# 安装依赖
go mod tidy

# 启动服务
go run ./cmd/server/

# 健康检查
curl http://localhost:8080/api/v1/schools
```

默认端口 `8080`，配置文件 `config.yaml`。

---

## API 文档

### 通用说明

- **Base URL:** `http://localhost:8080`
- **Content-Type:** `application/json`
- **响应格式:**

```json
{
  "code": 200,
  "message": "success",
  "data": { }
}
```

列表接口使用 `data.items` + `data.total`：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 4,
    "items": [ ]
  }
}
```

---

### 1. 学校信息

#### `GET /api/v1/schools` — 获取学校列表

返回所有已启用的学校。

<details>
<summary>测试结果</summary>

```bash
curl http://localhost:8080/api/v1/schools
```

**响应 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 4,
    "items": [
      {
        "code": "XAUAT",
        "name": "西安建筑科技大学",
        "website": "https://www.xauat.edu.cn",
        "features": ["login", "timetable", "grade_query", "exam_schedule", "..."],
        "enabled": true,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```
</details>

#### `GET /api/v1/schools/:code` — 获取学校详情

<details>
<summary>测试结果</summary>

```bash
curl http://localhost:8080/api/v1/schools/XAUAT
```

**响应 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "code": "XAUAT",
    "name": "西安建筑科技大学",
    "website": "https://www.xauat.edu.cn",
    "features": ["login", "timetable", "grade_query", "gpa_calculation", "exam_schedule", "semester_info", "bus_schedule", "program", "study_progress"],
    "enabled": true,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```
</details>

**错误码:**

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 404 | 学校不存在 |

---

### 2. XAUAT — 教务功能

所有 XAUAT 接口挂载在 `/api/v1/schools/XAUAT/` 下。

#### 认证方式

客户端先调用 `/login` 获取 Cookie，后续请求通过以下两种方式之一传递：

- **Cookie header:** `Cookie: __pstsid__=...; SESSION=...`
- **xauat header:** `xauat: __pstsid__=...; SESSION=...`

---

#### `POST /api/v1/schools/XAUAT/login` — SSO 登录

| 参数 | 位置 | 必填 | 说明 |
|------|------|------|------|
| username | body | 是 | 学号 |
| password | body | 是 | 密码 |

<details>
<summary>测试结果</summary>

**请求:**
```bash
curl -X POST http://localhost:8080/api/v1/schools/XAUAT/login \
  -H "Content-Type: application/json" \
  -d '{"username":"2201010101","password":"123456"}'
```

**成功 200:**
```json
{
  "code": 200,
  "message": "login success",
  "data": {
    "success": true,
    "student_id": "2201010101",
    "cookie": "__pstsid__=abc123; SESSION=xyz789"
  }
}
```

**参数缺失 400:**
```json
{
  "code": 400,
  "message": "invalid request: username and password required",
  "data": null
}
```

**凭证错误 401:**
```json
{
  "code": 401,
  "message": "login failed: invalid credentials",
  "data": null
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/courses` — 课表查询

| 参数 | 必填 | 说明 |
|------|------|------|
| student_id | 是 | 学号，支持逗号分隔多 ID |
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/courses?student_id=2201010101" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "data": [
      {
        "week_indexes": [1,2,3,4,5,6,7,8],
        "teachers": ["张三"],
        "campus": "雁塔校区",
        "room": "主楼301",
        "course_name": "高等数学",
        "course_code": "MATH101",
        "weekday": 1,
        "start_unit": 1,
        "end_unit": 2,
        "credits": "4",
        "lesson_id": "L001"
      }
    ],
    "expiration_time": "2026-05-16T00:00:00Z"
  }
}
```

**缺少参数 400:** `{"code":400, "message":"student_id is required"}`

**无 Cookie 401:** `{"code":401, "message":"cookie is required"}`
</details>

---

#### `GET /api/v1/schools/XAUAT/scores` — 成绩查询

| 参数 | 必填 | 说明 |
|------|------|------|
| student_id | 是 | 学号 |
| semester | 是 | 学期标识，如 `301` |
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/scores?student_id=2201010101&semester=301" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "高等数学",
      "lesson_code": "MATH101",
      "lesson_name": "高等数学(上)",
      "grade": "85",
      "gpa": "3.5",
      "grade_detail": "期末: 85; 平时: 90",
      "credit": "4",
      "is_minor": false
    }
  ]
}
```

**缺少参数 400:** `{"code":400, "message":"student_id and semester are required"}`
</details>

---

#### `GET /api/v1/schools/XAUAT/scores/semesters` — 学期列表

| 参数 | 必填 | 说明 |
|------|------|------|
| student_id | 否 | 学号 |
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/scores/semesters" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "data": [
      {"value": "301", "text": "2025-2026-1"},
      {"value": "302", "text": "2025-2026-2"}
    ]
  }
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/scores/current-semester` — 当前学期

| 参数 | 必填 | 说明 |
|------|------|------|
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/scores/current-semester" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "value": "302",
    "text": "2025-2026-2"
  }
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/exams` — 考试安排

| 参数 | 必填 | 说明 |
|------|------|------|
| student_id | 是 | 学号 |
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/exams?student_id=2201010101" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "exams": [
      {
        "name": "高等数学",
        "time": "2026-01-15 09:00-11:00",
        "location": "主楼301",
        "seat": "12"
      }
    ],
    "can_click": true
  }
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/bus` — 校车时刻表

无需认证。

| 参数 | 必填 | 说明 |
|------|------|------|
| date | 否 | 日期，默认当天 |
| loc | 否 | 线路，默认 `ALL` |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/bus?date=2026-05-09&loc=ALL"
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "records": [
      {
        "line_name": "草堂-雁塔",
        "description": "通勤班车",
        "departure_station": "草堂校区",
        "arrival_station": "雁塔校区",
        "run_time": "07:30",
        "arrival_station_time": "08:30"
      }
    ],
    "total": 1
  }
}
```

也支持路径参数：`GET /api/v1/schools/XAUAT/bus/CaCao`
</details>

---

#### `GET /api/v1/schools/XAUAT/program` — 培养方案

| 参数 | 必填 | 说明 |
|------|------|------|
| id | 是 | 培养方案 ID |
| name | 否 | 按课程名过滤 |
| dict | 否 | `"true"` 时返回树形结构 |
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/program?id=ROOT123" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "高等数学",
      "lesson_type": "必修",
      "exam_mode": "考试",
      "course_type_name": "公共基础课",
      "credits": 4.0,
      "term_str": "2025-2026-1"
    }
  ]
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/info/completion` — 学业进度

| 参数 | 必填 | 说明 |
|------|------|------|
| Cookie / xauat | 是 | 登录凭据 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/info/completion" \
  -H "Cookie: __pstsid__=...; SESSION=..."
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "type": "必修",
      "total": {"name": "总计", "actual": 120, "full": 140},
      "other": [
        {"name": "公共基础", "actual": 40, "full": 45}
      ]
    }
  ]
}
```
</details>

---

#### `GET /api/v1/schools/XAUAT/info/time` — 学期时间范围

Cookie 可选。提供有效 Cookie 时会从当前学期推算日期，否则回退到配置文件中的值。

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/info/time"
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "start_time": "2026-03-01",
    "end_time": "2026-07-18"
  }
}
```

传入 Cookie 时，会根据抓取的学期文本（如 `"2025-2026-2"`）动态计算日期：
- 第 1 学期：`09-01` ~ `01-15`
- 第 2 学期：`02-25` ~ `07-15`
</details>

---

#### `GET /api/v1/schools/XAUAT/payment/:id` — 校园卡登录

获取支付系统 Bearer Token。缓存 1 小时。

| 参数 | 必填 | 说明 |
|------|------|------|
| id (路径) | 是 | 校园卡号 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/payment/123456"
```

**成功 200:**
```json
{
  "code": 200,
  "message": "login success",
  "data": "eyJhbGciOiJIUzI1NiIs..."
}
```

**外部 API 不可达 503:**
```json
{
  "code": 503,
  "message": "keyboard request failed: ...",
  "data": null
}
```

> 此接口依赖校内支付系统 `ydfwpt.xauat.edu.cn`，需在校内网络环境下使用。
</details>

---

#### `GET /api/v1/schools/XAUAT/payment/:id/turnover` — 消费记录

获取最近 8 笔消费记录和电子账户余额。缓存 20 分钟。

| 参数 | 必填 | 说明 |
|------|------|------|
| id (路径) | 是 | 校园卡号 |

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/schools/XAUAT/payment/123456/turnover"
```

**成功 200:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "records": [
      {
        "turnover_type": "消费",
        "datetime_str": "2026-05-09 12:00:00",
        "resume": "食堂消费",
        "tranamt": 12.50
      }
    ],
    "total": 86.30
  }
}
```
</details>

---

### 3. 管理员接口

所有接口挂载在 `/api/v1/admin/` 下，需要认证。

```
Authorization: Bearer <admin_token>
```

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/schools` | 列出所有学校（含未启用） |
| POST | `/api/v1/admin/schools` | 新增学校 |
| PUT | `/api/v1/admin/schools/:code` | 更新学校（部分更新） |
| DELETE | `/api/v1/admin/schools/:code` | 删除学校 |

#### `GET /api/v1/admin/schools`

<details>
<summary>测试结果</summary>

```bash
curl "http://localhost:8080/api/v1/admin/schools" \
  -H "Authorization: Bearer luminous-admin-secret-token"
```

**成功 200:** 返回所有学校（包含 `enabled: false` 的项）

**无认证 401:**
```json
{"code":401, "message":"missing authorization header", "data":null}
```

**格式错误 401:**
```json
{"code":401, "message":"invalid authorization format, expected: Bearer <token>", "data":null}
```

**Token 错误 401:**
```json
{"code":401, "message":"invalid token", "data":null}
```
</details>

#### `POST /api/v1/admin/schools`

<details>
<summary>测试结果</summary>

```bash
curl -X POST "http://localhost:8080/api/v1/admin/schools" \
  -H "Authorization: Bearer luminous-admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TEST",
    "name": "测试大学",
    "website": "https://test.edu.cn",
    "features": ["timetable", "grade_query"]
  }'
```

**成功 201:**
```json
{
  "code": 201,
  "message": "school created",
  "data": {
    "code": "TEST",
    "name": "测试大学",
    "website": "https://test.edu.cn",
    "features": ["timetable", "grade_query"],
    "enabled": true,
    "created_at": "2026-05-09T15:09:44+08:00",
    "updated_at": "2026-05-09T15:09:44+08:00"
  }
}
```

**重复 Code 409:**
```json
{"code":409, "message":"school with code TEST already exists", "data":null}
```

**无效 Feature 400:**
```json
{"code":400, "message":"invalid feature: xxx", "data":null}
```
</details>

#### `PUT /api/v1/admin/schools/:code`

<details>
<summary>测试结果</summary>

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/schools/TEST" \
  -H "Authorization: Bearer luminous-admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "新名称", "enabled": false}'
```

**成功 200:**
```json
{
  "code": 200,
  "message": "school updated",
  "data": { "...", "name": "新名称", "enabled": false }
}
```

**不存在 404:** `{"code":404, "message":"school not found"}`
</details>

#### `DELETE /api/v1/admin/schools/:code`

<details>
<summary>测试结果</summary>

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/schools/TEST" \
  -H "Authorization: Bearer luminous-admin-secret-token"
```

**成功 200:**
```json
{"code":200, "message":"school deleted", "data":null}
```
</details>

---

### 4. 指标监控

#### `GET /metrics` — Prometheus 指标

无需认证。

<details>
<summary>测试结果</summary>

```bash
curl http://localhost:8080/metrics
```

**自定义指标:**

```
# HELP luminous_http_requests_total Total number of HTTP requests.
# TYPE luminous_http_requests_total counter
luminous_http_requests_total{method="GET",path="/api/v1/schools",status="200"} 1
luminous_http_requests_total{method="POST",path="/api/v1/schools/XAUAT/login",status="400"} 1

# HELP luminous_http_request_duration_seconds HTTP request duration in seconds.
# TYPE luminous_http_request_duration_seconds histogram
luminous_http_request_duration_seconds_bucket{method="GET",path="/api/v1/schools",le="0.005"} 1
...
```

同时暴露 Go runtime、进程、GC 等标准指标。
</details>

---

### 5. 错误码速查

| 状态码 | 场景 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 参数缺失 / 格式错误 |
| 401 | 未认证 (Cookie 缺失 / Token 无效) |
| 404 | 学校不存在 |
| 409 | 资源冲突 (重复创建) |
| 500 | 服务器内部错误 |
| 503 | 外部服务不可用 (支付系统等) |

---

## 配置

`config.yaml`:

```yaml
server:
  port: 8080
  mode: debug          # debug | release | test

auth:
  admin_token: "luminous-admin-secret-token"

data:
  schools_file: "./data/schools.json"

schools:
  xauat:
    base_url: "https://swjw.xauat.edu.cn"
    login_url: "https://schedule.xauat.site"
    old_bus_url: "https://school-bus.xauat.edu.cn"
    new_bus_url: "https://bcdd.xauat.edu.cn"
    semester_start: "2026-03-01"
    semester_end: "2026-07-18"
```

支持环境变量覆盖，前缀 `LUMINOUS_`：
- `LUMINOUS_SERVER_PORT=9090`
- `LUMINOUS_AUTH_ADMIN_TOKEN=xxx`

## 项目结构

```
Luminous/
├── cmd/server/main.go                  # 入口
├── config.yaml                         # 配置文件
├── data/schools.json                   # 学校数据（JSON 文件存储）
├── internal/
│   ├── config/config.go                # 配置加载
│   ├── model/school.go                 # 学校模型 + Feature 定义
│   ├── repository/
│   │   ├── school_repo.go              # JSON 文件仓库
│   │   └── school_repo_test.go
│   ├── handler/
│   │   ├── school.go                   # 学校查询接口
│   │   ├── admin.go                    # 管理员 CRUD
│   │   ├── school_service.go           # SchoolServiceHandler 接口
│   │   ├── xauat.go                    # XAUAT 所有路由注册
│   │   └── school_test.go              # Handler 测试
│   ├── middleware/
│   │   ├── auth.go                     # Bearer Token 认证
│   │   ├── cors.go                     # CORS 中间件
│   │   └── metrics.go                  # Prometheus 指标
│   ├── response/response.go            # 统一响应格式
│   ├── router/router.go                # 路由组装
│   ├── school/xauat/                   # XAUAT 业务逻辑
│   │   ├── models.go                   # 数据模型
│   │   ├── config.go                   # URL/日期配置 + Init()
│   │   ├── login.go                    # SSO 登录
│   │   ├── course.go                   # 课表抓取
│   │   ├── score.go                    # 成绩抓取
│   │   ├── exam.go                     # 考试安排
│   │   ├── bus.go                      # 校车时刻表
│   │   ├── program.go                  # 培养方案
│   │   ├── info.go                     # 学业进度 + 学期日期
│   │   ├── payment.go                  # 校园卡支付
│   │   ├── semester.go                 # 学期解析
│   │   └── shutdown.go                 # 缓存优雅关闭
│   └── util/
│       ├── httpclient.go               # HTTP 客户端（重试、UA 轮换）
│       └── cache.go                    # 内存 TTL 缓存
```

## 扩展新学校

实现 `SchoolServiceHandler` 接口即可注册：

```go
type SchoolServiceHandler interface {
    Code() string
    RegisterRoutes(rg *gin.RouterGroup)
}
```

然后在 `main.go` 中传入 `router.SetupRouter()`。
