# API Hub

本地 OpenAI 兼容 API 聚合代理。在 GUI 中选择厂商和模型，所有请求自动路由到选定目标，无需修改客户端代码。

## 功能

- Web GUI 管理界面，自动打开浏览器
- 选择 Provider 和 Model 后自动保存
- 所有 API 请求强制使用 GUI 选择的厂商和模型路由
- 客户端请求中的 `model` 字段不影响路由（写任意值即可）
- 支持 `/v1/chat/completions`、`/v1/embeddings`、`/v1/models`
- 支持流式 SSE 响应透传
- 一个二进制同时支持 GUI 模式和纯 API 模式

## 构建

```bash
go build -o api_hub.exe ./cmd/api_hub
```

## 启动

默认启动 Web GUI（自动打开浏览器）：

```bash
./api_hub.exe
```

仅启动 API 代理（无 GUI，需预先配置 `defaults`）：

```bash
./api_hub.exe -mode api
```

指定配置文件：

```bash
./api_hub.exe -config config.local.yaml
```

启动参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-mode` | `gui` | `gui` Web 管理页 / `api` 仅启动代理 |
| `-config` | `config.yaml` | 配置文件路径 |
| `-gui-listen` | `127.0.0.1:8090` | Web 管理页监听地址 |
| `-open-browser` | `true` | 启动 GUI 时自动打开浏览器 |

## Web GUI

启动后访问 `http://127.0.0.1:8090`（默认自动打开浏览器）。

- 选择 Provider → Model 下拉框自动更新为该厂商的模型列表
- 切换选择**自动保存**到 `config.yaml` 的 `defaults` 字段，无需手动保存
- 点击「启动服务」启动 API 代理，**下拉框锁定**，显示客户端连接信息
- 点击「停止服务」方可重新选择厂商和模型
- 客户端 `model` 字段写**任意值**均不影响路由

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
      - deepseek-v4-pro
```

- `server.address`：API 代理监听地址
- `server.api_key`：客户端访问本地 API 时使用的密钥
- `defaults.provider / defaults.model`：GUI 页面初始值，API 启动时的默认路由
- `providers.<name>.base_url`：上游 API 根地址
- `providers.<name>.api_key`：上游 API Key，支持 `${ENV_NAME}`
- `providers.<name>.models`：`/v1/models` 展示的模型列表和 GUI 下拉选项

## 使用方式

客户端设置：

- Base URL：`http://127.0.0.1:8080/v1`
- API Key：`config.yaml` 中 `server.api_key`

### Chat Completions

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anything",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

### 流式请求

```bash
curl http://127.0.0.1:8080/v1/chat/completions \
  -H "Authorization: Bearer local-dev-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anything",
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
    "model": "anything",
    "input": "hello"
  }'
```

## 路由规则

API 服务器启动后，**所有请求**强制路由到 GUI 选择的 provider 和 model。

- 客户端 JSON body 中 `model` 字段的值**不影响路由**，可写任意非空值
- 服务端会强制替换为 GUI 选中的模型，并转发到对应厂商
- 切换厂商/模型需在 GUI 中操作（先停止服务 → 重新选择 → 启动服务）

`models` 列表的用途：
1. 填充 GUI 下拉框选项
2. `/v1/models` 返回聚合模型列表
3. GUI 保存 defaults 时校验合法性

**不影响请求路由**。

## 常见错误

| 状态码 | 说明 |
|--------|------|
| 401 | 本地 `Authorization` 缺失或不匹配 |
| 400 no provider selected | 未在 GUI 中选择 provider 就启动了 API 服务 |
| 400 unknown provider | `defaults.provider` 不在 `providers` 中 |
| 400 method not allowed | 请求方法不对（如 GET chatbot 接口） |
| 502 | 本地代理无法连接上游 API |

## 开发验证

```bash
go test ./...
go build -o api_hub.exe ./cmd/api_hub
```
