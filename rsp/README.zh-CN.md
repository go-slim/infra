# 响应处理包 (rsp)

一个为 Go Web 应用程序设计的全面 HTTP 响应处理系统，提供结构化响应并支持多种内容类型和自动内容协商。

## 特性

- **自动内容协商**: 支持 JSON、JSONP、HTML、XML 和纯文本响应，基于 Accept 头部自动选择
- **结构化错误处理**: 丰富的错误报告，包含问题详情和字段特定的验证错误
- **RESTful 辅助函数**: 为常见 HTTP 状态响应提供预构建函数 (OK、Created、Deleted、Accepted)
- **函数选项模式**: 使用可组合选项进行灵活的响应配置
- **验证集成**: 与 go-slim.dev/v 验证库无缝集成
- **标准化响应格式**: 所有响应类型使用一致的 JSON 结构

## 安装

```go
import "go-slim.dev/infra/rsp"
```

## 快速开始

### 基本用法

```go
// 简单成功响应
rsp.Ok(c, userData)

// 创建响应带数据
rsp.Created(c, newResource)

// 错误响应带自定义选项
rsp.Respond(c,
    rsp.StatusCode(400),
    rsp.Message("输入无效"),
    rsp.Data(validationErrors),
)
```

### 响应格式

所有响应都遵循这个标准化结构：

```json
{
  "code": "SUCCESS",
  "ok": true,
  "msg": "OK",
  "data": {...},           // 可选
  "problems": {...},       // 可选，用于验证错误
  "error": "..."           // 可选，仅在调试模式下
}
```

## API 参考

### HTTP 响应辅助函数

#### `Ok(c slim.Context, data ...any) error`

使用 HTTP 200 状态码响应成功请求。

#### `Created(c slim.Context, data ...any) error`

使用 HTTP 201 状态码响应成功的资源创建。

#### `Deleted(c slim.Context, data ...any) error`

使用适当的 HTTP 状态码响应成功的资源删除。

- **带数据时**：返回 HTTP 200 (OK) 状态码，包含响应数据
- **无数据时**：返回 HTTP 204 (No Content) 状态码，空响应体

适用于删除确认或返回删除资源详情的场景。

#### `Accepted(c slim.Context, data ...any) error`

使用 HTTP 202 状态码响应已接受的异步操作。

### 配置选项

#### `StatusCode(status int) Option`

设置响应的 HTTP 状态码。

#### `Header(key, value string) Option`

向响应添加自定义 HTTP 头部。

#### `Cookie(cookie *http.Cookie) Option`

在响应中设置 HTTP cookie。

#### `Message(msg string) Option`

设置响应的自定义消息。

#### `Data(data any) Option`

设置响应的数据载荷。

### 错误处理

包通过 Problem 系统提供结构化错误报告：

```go
type Problem struct {
    Code     string   `json:"code"`
    Message  string   `json:"msg"`
    Problems Problems `json:"problems,omitempty"`
}

type Problems map[string][]*Problem
```

## 示例

### 带多个选项的自定义响应

```go
rsp.Respond(c,
    rsp.StatusCode(http.StatusCreated),
    rsp.Header("Location", "/api/users/123"),
    rsp.Header("X-API-Version", "1.0"),
    rsp.Message("用户创建成功"),
    rsp.Data(map[string]interface{}{
        "id": 123,
        "username": "张三",
    }),
)
```

### 带验证问题的错误响应

```go
problems := make(rsp.Problems)
problems.Add(&rsp.Problem{
    Label:   "email",
    Code:    "INVALID_FORMAT",
    Message: "邮箱格式无效",
})
problems.Add(&rsp.Problem{
    Label:   "password",
    Code:    "TOO_SHORT",
    Message: "密码至少需要8个字符",
})

rsp.Respond(c,
    rsp.StatusCode(http.StatusBadRequest),
    rsp.Message("验证失败"),
    rsp.Data(problems),
)
```

### 设置 Cookie

```go
sessionCookie := &http.Cookie{
    Name:     "session_id",
    Value:    "abc123def456",
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    MaxAge:   3600,
}

rsp.Ok(c, userData)
// 或者带自定义 cookie
rsp.Respond(c,
    rsp.Data(userData),
    rsp.Cookie(sessionCookie),
)
```

## 支持的内容类型

包根据 `Accept` 头部自动协商响应内容：

- **JSON**: `application/json`
- **JSONP**: `application/javascript` (带回调)
- **HTML**: `text/html`
- **XML**: `application/xml`
- **Text**: `text/plain`, `text/*`

## 配置

### 自定义序列化

您可以自定义 HTML 和文本序列化：

```go
rsp.HTMLMarshaller = func(data map[string]any) (string, error) {
    // 自定义 HTML 渲染逻辑
    return renderTemplate("response", data), nil
}

rsp.TextMarshaller = func(data map[string]any) (string, error) {
    // 自定义文本格式化逻辑
    return formatAsText(data), nil
}
```

### JSONP 配置

配置 JSONP 回调参数名：

```go
rsp.JsonpCallbacks = []string{"callback", "cb", "jsonp"}
rsp.DefaultJsonpCallback = "callback"
```

## 验证集成

包与 `go-slim.dev/v` 验证库无缝集成：

```go
// 将验证错误转换为问题
problems := make(rsp.Problems)
for _, err := range validationErrors.All() {
    problems.AddError(err)
}

rsp.Respond(c,
    rsp.StatusCode(http.StatusBadRequest),
    rsp.Data(problems),
)
```

## 许可证

此包是 go-slim/infra 项目的一部分。
