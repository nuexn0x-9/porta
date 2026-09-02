# PORTA CONFIGURATION REFERENCE (`porta.yaml`)

This document is the official reference for the `porta.yaml` configuration schema.

---

## 1. Complete Schema Example

```yaml
# porta.yaml
version: "1"

project:
  name: fullstack-store         # Required: Lowercase alphanumeric, hyphen, underscore
  environment: development      # Optional: development | staging | testing (default: development)

services:
  web:
    host: 127.0.0.1             # Optional: Target host (default: 127.0.0.1)
    port: 3000                  # Required: Port number (1-65535)
    route: /                    # Optional: Ingress URL prefix (default: /)
    websocket: true             # Optional: Enable WebSocket upgrade (default: true)

  api:
    host: 127.0.0.1
    port: 8000
    route: /api
    strip_path: true            # Optional: Strip '/api' prefix before forwarding (default: false)
    health_check:
      path: /healthz            # Optional: HTTP probe path (default: TCP ping if empty)
      interval: 5s              # Optional: Probe interval (default: 5s)
      timeout: 2s               # Optional: Probe timeout (default: 2s)

tunnel:
  provider: cloudflare          # Optional: Tunnel provider (default: cloudflare)

security:
  mode: password                # Optional: public | password | token (default: public)
  password: ${AUTH_PASS:-admin:SecretPass123!} # Required if mode=password
  token: ${API_TOKEN}           # Required if mode=token
```

---

## 2. Field Reference Table

| Key Path | Type | Required | Default | Description & Validation Rules |
| :--- | :--- | :--- | :--- | :--- |
| `version` | string | No | `"1"` | Configuration schema version. |
| `project.name` | string | **Yes** | None | Project identifier. Regex: `^[a-zA-Z0-9_-]+$`. |
| `project.environment`| string | No | `"development"`| Metadata environment label. |
| `services` | map | **Yes** | None | Map of services. Minimum 1 service required (`len >= 1`). |
| `services.<id>.port` | integer| **Yes** | None | Local target port (`1` to `65535`). |
| `services.<id>.host` | string | No | `"127.0.0.1"` | Target host. Must be loopback (`127.0.0.1`, `localhost`, `::1`). |
| `services.<id>.route`| string | No | `/` (single) | Ingress path prefix. Must start with `/`. Must be unique. |
| `services.<id>.strip_path`| boolean| No | `false` | If `true`, strips the matching prefix before forwarding. |
| `services.<id>.websocket` | boolean| No | `true` | Enables transparent WebSocket connection hijacking. |
| `services.<id>.health_check.path` | string | No | `""` | HTTP GET health path (e.g. `/healthz`). Defaults to TCP socket ping if empty. |
| `services.<id>.health_check.interval` | duration | No | `5s` | Time between consecutive health probes. |
| `services.<id>.health_check.timeout` | duration | No | `2s` | Timeout for each health check probe. |
| `tunnel.provider` | string | No | `"cloudflare"`| Egress tunnel provider. |
| `security.mode` | string | No | `"public"` | Ingress protection: `public`, `password`, or `token`. |
| `security.password` | string | Conditional | `""` | Required if `mode: password`. Format: `user:pass` or `pass`. |
| `security.token` | string | Conditional | `""` | Required if `mode: token`. Matches Bearer token or `?porta_token=`. |

---

## 3. Environment Variable Expansion

PORTA supports dynamic environment variable substitution using bash-like syntax:
- `${VARIABLE_NAME}`
- `${VARIABLE_NAME:-default_fallback_value}`

### Example:
```yaml
security:
  mode: password
  password: ${PORTA_AUTH_PASSWORD:-admin:ChangeMeInProduction!}
```

---

## 4. Path Stripping Behavior (`strip_path`)

| Inbound Public Request | Service Route | `strip_path` | Target Dispatched URI |
| :--- | :--- | :--- | :--- |
| `GET /api/v1/users` | `/api` | `false` | `GET http://127.0.0.1:8000/api/v1/users` |
| `GET /api/v1/users` | `/api` | `true` | `GET http://127.0.0.1:8000/v1/users` |
| `GET /about` | `/` | `false` | `GET http://127.0.0.1:3000/about` |

---
*End of Configuration Reference.*
