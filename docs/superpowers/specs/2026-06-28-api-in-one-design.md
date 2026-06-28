# API-in-One Design

Date: 2026-06-28

## Goal

Build a local Go service that aggregates multiple OpenAI-compatible APIs behind one OpenAI-compatible API. Clients use one local base URL and one local API key, then select the upstream provider by using a model prefix.

## Scope

The first version supports:

- `/v1/chat/completions`, including streaming responses
- `/v1/embeddings`
- `/v1/models`
- Transparent forwarding for other `/v1/*` endpoints
- YAML-based provider configuration
- Local bearer-token authentication
- Provider selection by model prefix, for example `openai/gpt-4o` or `deepseek/deepseek-chat`

The first version does not include a Web UI, persistent database, usage dashboard, load balancing, or automatic failover.

## Configuration

The app reads `config.yaml` from the working directory by default. The config defines the server address, local API key, and providers.

Example shape:

```yaml
server:
  address: 127.0.0.1:8080
  api_key: local-dev-key

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

Environment variable references in provider keys use `${NAME}` syntax and are resolved at startup.

## Routing

For request bodies that contain `model`, the value must use `provider/model-name`.

Example request model:

```json
{
  "model": "deepseek/deepseek-chat"
}
```

The router selects provider `deepseek`, rewrites `model` to `deepseek-chat`, and forwards the request to the provider base URL.

For endpoints without a model field, transparent forwarding requires a provider prefix in the path or query in a later version. In the first version, `/v1/models` is handled locally and other model-less `/v1/*` requests return a clear error unless they can be routed from the request body.

## Components

- Config loader: reads YAML, resolves environment variables, validates provider names, base URLs, and keys.
- Auth middleware: requires `Authorization: Bearer <server.api_key>` on all `/v1/*` requests.
- Router: parses OpenAI-compatible request JSON when needed and selects the upstream provider.
- Proxy: forwards method, path, headers, and body to the selected provider with the provider API key.
- Models handler: returns a merged OpenAI-compatible model list where each model id is prefixed with the provider name.
- Error writer: returns OpenAI-style JSON errors for auth, config, routing, and upstream failures.

## Data Flow

1. Client sends a request to local `/v1/...` with the local bearer token.
2. Auth middleware validates the token.
3. Router reads the request body when routing needs the `model` field.
4. Router extracts provider and upstream model from `provider/model`.
5. Proxy rebuilds the request body with the stripped model name.
6. Proxy sends the request to `provider.base_url + original_path` using the provider API key.
7. Response status, headers, and body are streamed back to the client.

## Error Handling

- Missing or invalid local token returns `401`.
- Unknown provider prefix returns `400`.
- Missing provider prefix in a route that needs model routing returns `400`.
- Invalid config fails startup with a descriptive message.
- Upstream network errors return `502`.
- Upstream response bodies are passed through unchanged when the upstream returns a response.

## Testing

Tests should cover:

- Config loading and environment variable resolution
- Local bearer-token authentication
- Model prefix parsing and body rewrite
- `/v1/models` aggregation
- Proxy forwarding behavior with `httptest` upstream servers
- Streaming response passthrough
- Error responses for missing token, unknown provider, and missing model prefix
