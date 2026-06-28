# Web GUI Design

Date: 2026-06-28

## Goal

Add a simple local Web GUI for managing `api-in-one`: choosing a default API provider/model, saving that choice to `config.yaml`, and starting or stopping the API proxy service from the same executable.

## Scope

The GUI version supports:

- A Web UI served at `http://127.0.0.1:8090` by the main `api-in-one` executable in GUI mode.
- A single executable: `api-in-one` supports both GUI mode and API-only mode.
- Automatically opening the GUI URL in the system browser after startup when `-open-browser=true`.
- Reading `config.yaml` from the working directory by default.
- Listing configured providers and their models.
- Selecting a default provider and default model.
- Automatically writing the selected default to `config.yaml` when the provider or model dropdown changes.
- Starting the API proxy inside the same process from the GUI.
- Stopping the API proxy started by the GUI.
- Locking provider/model selection while the API server is running.
- Showing the client connection values while running: Base URL, API Key, and Model.
- Showing basic service status and recent logs.
- A cleaner, more polished Web UI with a modern card layout, clearer status indicators, copyable client values, and disabled-state styling.
- Default routing in the API proxy: if a request model has no `provider/` prefix, route it to the configured default provider/model.

The first version does not include provider editing, key editing, system tray integration, background installation, multi-user auth for the GUI, or managing API processes that were not started by the GUI.

## Configuration

Extend `config.yaml` with a `defaults` section:

```yaml
defaults:
  provider: openrouter
  model: qwen/qwen3-coder:free
```

The existing provider configuration remains unchanged. The GUI writes only the `defaults` section when saving the selected API.

The default model is stored as the upstream model name exactly as it appears in that provider's `models` list. If users choose provider `openrouter` and model `qwen/qwen3-coder:free`, clients may send either:

```json
{"model":"openrouter/qwen/qwen3-coder:free"}
```

or, when default routing is desired:

```json
{"model":"qwen/qwen3-coder:free"}
```

For unprefixed model requests, the proxy routes to `defaults.provider` and rewrites `model` to `defaults.model`.

## Commands

Keep one executable with mode selection:

```bash
api-in-one -mode gui -config config.yaml -gui-listen 127.0.0.1:8090 -open-browser=true
```

The default mode is GUI mode, so this is enough for normal local use:

```bash
api-in-one
```

API-only mode remains available:

```bash
api-in-one -mode api -config config.yaml
```

Startup flags:

- `-mode`: defaults to `gui`; valid values are `gui` and `api`.
- `-config`: defaults to `config.yaml`.
- `-gui-listen`: defaults to `127.0.0.1:8090`.
- `-open-browser`: defaults to `true` in GUI mode.
- API listen address is read from `server.address` in the config. If it is empty, config loading defaults it to `127.0.0.1:8080`.

## GUI Layout

The page has four areas:

1. Header: title, selected config file, API service status.
2. Provider selector: dropdown for provider and dropdown/list for that provider's models. These controls are enabled only when the API server is stopped.
3. Controls: Start Service, Stop Service, Refresh.
4. Client info panel: visible while running, showing copyable Base URL `http://<server.address>/v1`, API Key from `server.api_key`, and Model from the saved default model.
5. Logs/status panel: shows service lifecycle messages and errors.

The UI is plain HTML/CSS/JavaScript served by Go. No frontend build tool is required. Dropdown changes call `POST /api/defaults` immediately; there is no separate save button. The page should use a polished local-dashboard style: gradient background, cards, clear button hierarchy, status badges, disabled controls while running, and monospace copy fields for client values.

## Components

- Config extensions: add `DefaultsConfig` to the existing config package.
- Config persistence: load YAML while preserving provider values and write the selected defaults back to disk.
- Default routing: update router/server flow so unprefixed models can use configured defaults.
- Runtime API manager: starts/stops an in-process `http.Server` for the API proxy and tracks running state.
- Browser opener: opens the GUI URL using OS-specific commands when enabled.
- GUI server: serves HTML and JSON endpoints.

## GUI Endpoints

- `GET /` returns the HTML page.
- `GET /api/config` returns providers, models, defaults, process status, and client connection values.
- `POST /api/defaults` accepts `{ "provider": "...", "model": "..." }`, validates it, saves to config, and returns updated state. It is rejected while the API server is running.
- `POST /api/service/start` starts the in-process API server if not already running.
- `POST /api/service/stop` stops the in-process API server if running.
- `GET /api/service/status` returns running state, recent logs, and client connection values.

## Data Flow

1. User starts `api-in-one` with no arguments.
2. The executable enters GUI mode, starts the local GUI HTTP server, and opens the GUI URL in the system browser when `-open-browser=true`.
3. GUI loads `config.yaml` and displays provider/model choices.
4. User selects a provider or model.
5. GUI validates that provider and model exist in config and immediately writes `defaults.provider` and `defaults.model` to `config.yaml`.
6. User clicks Start Service.
7. GUI starts the API server in the same process using the same config file.
8. Provider/model selectors become disabled while the API server is running.
9. GUI displays the client Base URL, API Key, and Model to use in OpenAI-compatible clients.
10. API requests with prefixed models keep existing behavior.
11. API requests with unprefixed models use the saved defaults.
12. User clicks Stop Service to stop the API server and re-enable provider/model selection.

## Error Handling

- Invalid config returns a visible GUI error.
- Selecting an unknown provider/model returns `400` from GUI API.
- Changing provider/model while the service is running is blocked in the UI; the server also rejects `POST /api/defaults` while the API server is running.
- Browser auto-open failures are logged but do not stop the GUI server.
- Starting when already running returns current running status without creating another API server.
- Stopping when not running returns stopped status.
- API server start/stop lifecycle messages are captured and shown in the logs panel.
- API default routing returns `400` if no defaults are configured and the model is unprefixed.

## Testing

Tests should cover:

- Config loading with `defaults`.
- Saving defaults to YAML.
- Rejecting invalid default provider/model selections.
- Rejecting default changes while the GUI-managed API server is running.
- Unprefixed model routing through configured defaults.
- Existing prefixed model routing remains unchanged.
- GUI JSON endpoints with `httptest`.
- GUI state response includes client Base URL, API Key, and Model.
- Runtime API manager start/stop behavior using `httptest` or a loopback listener.
- Browser opener command construction is tested without launching a real browser.
- Main command flag defaults: `-mode gui`, `-config config.yaml`, `-gui-listen 127.0.0.1:8090`, `-open-browser true`.
