# PORTA CONFIGURATION GUIDE

This document provides an overview of configuring PORTA using `porta.yaml`. For the full schema specification, see [CONFIGURATION_REFERENCE.md](CONFIGURATION_REFERENCE.md).

---

## 1. Minimal Single-Service Configuration

Exposing a single frontend or backend application (e.g. `localhost:3000`):

```yaml
version: "1"

project:
  name: my-app

services:
  web:
    port: 3000
    route: /
```

---

## 2. Multi-Service Configuration

Exposing both a frontend application and a backend REST/GraphQL API under a single public domain:

```yaml
version: "1"

project:
  name: my-fullstack-app

services:
  frontend:
    port: 3000
    route: /

  backend:
    port: 8000
    route: /api
    strip_path: false
    health_check:
      path: /healthz
      interval: 5s

security:
  mode: password
  password: ${PORTA_PASSWORD:-admin:SecretPass123!}
```

---

## 3. Configuration Best Practices
1. **Never Hardcode Secrets:** Use environment variable expansion (`${AUTH_TOKEN}`) in `porta.yaml`.
2. **Use Unique Routes:** Every service route prefix must be unique (e.g., `/`, `/api`, `/ws`) to prevent route collisions.
3. **Enable Health Checks:** Provide HTTP health check endpoints (`/healthz` or `/status`) so PORTA can provide 502 error pages if a service is down.

---
