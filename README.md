# api-in-one

本地 OpenAI 兼容 API 聚合代理。客户端只需要连接一个本地地址，通过 `provider/model` 格式切换不同上游 OpenAI-compatible API。

## 功能

- 统一本地 OpenAI-compatible API 地址
- 使用一个本地 `Authorization: Bearer <api_key>` 保护服务
- 通过模型名前缀路由：`openai/gpt-4o`、`deepseek/deepseek-chat`
- 支持 `/v1/models` 聚合模型列表
- 支持 POST `/v1/*` 请求转发，包括 `/v1/chat/completions`、`/v1/embeddings`
- 支持流式 SSE 响应透传
- 使用 `config.yaml` 配置多个上游服务

## 构建

```bash
go build ./cmd/api-in-one
```

Windows 生成：

```text
api-in-one.exe
```

## 启动

默认启动 Web GUI（自动打开浏览器）：

```bash
./api-in-one
```

仅启动 API 代理（无 GUI）：

```bash
./api-in-one -mode api
```

指定配置文件：

```bash
./api-in-one -config config.yaml
```

启动参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-mode` | `gui` | `gui` Web 管理页 / `api` 仅启动代理 |
| `-config` | `config.yaml` | 配置文件路径 |
| `-gui-listen` | `127.0.0.1:8090` | Web 管理页监听地址 |
| `-open-browser` | `true` | 启动 GUI 时自动打开浏览器 |

## Web GUI

启动后自动打开 `http://127.0.0.1:8090`。

- 选择 Provider 和 Model，自动保存到 `config.yaml` 的 `defaults` 字段
- 点击「启动服务」在同进程启动 API 代理，下拉框锁定
- 运行时显示客户端连接信息：Base URL、API Key、Model
- 点击「停止服务」方可重新选择厂商和模型
- 不选模型前缀时，自动使用 GUI 保存的默认 provider/model 路由

```yaml
defaults:
  provider: openai
  model: gpt-4o
```

## 配置

```yaml
server:
  address: 127.0.0.1:8080
  api_key: local-dev-key

defaults:
  provider: openai
  model: gpt-4o

providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    models:
      - gpt-4o
      - gpt-4o-mini
  deepseek:
    base_url: https://api.deepseek.com/v1
    api_key: ${DEEPSEEK_API_KEY}
    models:
      - deepseek-chat
```

- `server.address`：API 代理监听地址
- `server.api_key`：客户端访问本地 API 时使用的密钥
- `defaults.provider / defaults.model`：无 provider 前缀时的默认路由，GUI 自动管理
- `providers.<name>.base_url`：上游 API 根地址，通常以 `/v1` 结尾
- `providers.<name>.api_key`：上游 API Key，支持 `${ENV_NAME}`
- `providers.<name>.models`：`/v1/models` 展示的模型列表

## 使用方式


客户端 base URL 设置为：

```text
http://127.0.0.1:8080/v1
```

API Key 设置为 `config.yaml` 里的 `server.api_key`，例如：

```text
local-dev-key
```

### 查看模型

```bash
curl http://127.0.0.1:8080/v1/models \
  -H "Authorization: Bearer local-dev-key"
```

返回的模型 ID 会带 provider 前缀，例如：

```text
openai/gpt-4o
deepseek/deepseek-chat
```

### Chat Completions

显式切换厂商：

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-v4-pro",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

或走当前 GUI 默认（设好默认后直接用）：

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

代理行为：

1. 读取 `model` 字段
2. 带 `/`：取前缀为 provider，其余为 model 发上游
3. 不带 `/`：走 `defaults.provider` + `defaults.model`

### 流式请求

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o",
    "stream": true,
    "messages": [
      {"role": "user", "content": "讲个短笑话"}
    ]
  }'
```

### Embeddings

```bash
curl http://127.0.0.1:8080/v1/embeddings \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/text-embedding-3-small",
    "input": "hello"
  }'
```

注意：需要在对应 provider 的 `models` 列表中加入模型名才能从 `/v1/models` 看到。

## 路由规则

所有 POST `/v1/*` 请求通过 JSON body 中的 `model` 字段路由。

### 规则 1：带 Provider 前缀（显式路由）

```json
{"model": "deepseek/deepseek-v4-pro"}
```

- 按 `provider-name/upstream-model` 解析，`/` 之前为 provider name，之后为发给上游的 model 名
- 必须匹配 `config.yaml` 中 `providers.` 下的某个 key
- 无论 `defaults` 是否配置都生效
- model 名**不做校验**，直接透传给上游

### 规则 2：无 Provider 前缀（默认路由）

```json
{"model": "gpt-4o"}
```

- 使用 `config.yaml` 中 `defaults.provider` 和 `defaults.model` 路由
- 请求里的原 `model` 值**被忽略**，上游收到的是 `defaults.model`
- 如果 `defaults` 未配置或不全，返回 `400`

### 规则 3：`models` 字段

每个 provider 的 `models` 列表**只用于**：
1. `/v1/models` 返回聚合模型列表
2. GUI 自动保存默认 provider/model 时校验合法性

**不影响请求路由**。带前缀的请求即使 model 不在 `models` 列表里也会正常转发。

### 规则 4：请求 Body 要求

所有 POST `/v1/*` 请求必须在 JSON body 中包含 `model` 字段（字符串类型），否则返回 `400 request body must include model`。

## 常见错误

### 401 Unauthorized

本地 `Authorization` 缺失或不匹配。

```text
Authorization: Bearer local-dev-key
```

### 400 unknown provider

请求中的模型前缀没有出现在 `config.yaml` 的 `providers` 中。

### 400 model must use provider/model format

`model` 没有使用 `provider/model` 格式。

### 502 api_error

本地代理无法连接上游 API，通常是网络、`base_url` 或上游服务不可用。

## 开发验证

```bash
go test ./...
go build ./cmd/api-in-one
```
