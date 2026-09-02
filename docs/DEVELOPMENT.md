# PORTA DEVELOPER & CONTRIBUTOR GUIDE

This guide provides instructions for building, running, testing, and contributing to the **PORTA** codebase.

---

## 1. Prerequisites & Environment Setup

- **Language:** Go 1.22+ (tested through Go 1.27)
- **Dependencies:** Minimal third-party dependencies:
  - `github.com/spf13/cobra` (CLI framework)
  - `gopkg.in/yaml.v3` (YAML unmarshaler)
- **Supported Operating Systems:** Windows, macOS, Linux.

---

## 2. Repository Layout

```text
g:/PORTA/
├── cmd/
│   └── porta/
│       └── main.go                 # Binary entry point
├── internal/
│   ├── cli/                        # Cobra CLI command handlers
│   ├── config/                     # YAML parser, validator & model
│   ├── registry/                   # In-memory service registry
│   ├── router/                     # Longest Prefix Match (LPM) router
│   ├── proxy/                      # Reverse proxy, websocket & HTTP transport
│   ├── security/                   # SSRF guard, Basic Auth & Token Auth
│   ├── health/                     # Health checker engine
│   ├── tunnel/                     # Tunnel abstraction & Cloudflare driver
│   ├── ui/                         # Terminal UI & structured logger
│   └── utils/                      # Cross-platform process and net helpers
├── docs/                           # Architecture, audit, and user documentation
├── examples/                       # Example configurations
├── go.mod
└── go.sum
```

---

## 3. Build Commands

### Compile Local Executable:
```bash
# Windows
go build -o porta.exe ./cmd/porta

# Linux / macOS
go build -o porta ./cmd/porta
```

### Cross-Compilation:
```bash
# Build for Linux x64
GOOS=linux GOARCH=amd64 go build -o dist/porta-linux-amd64 ./cmd/porta

# Build for macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dist/porta-darwin-arm64 ./cmd/porta

# Build for Windows x64
GOOS=windows GOARCH=amd64 go build -o dist/porta-windows-amd64.exe ./cmd/porta
```

---

## 4. Testing & Verification

### Run Complete Unit & Integration Test Suite:
```bash
go test ./... -v
```

### Run Static Analysis:
```bash
go vet ./...
```

### Run Performance Benchmarks:
```bash
go test ./internal/proxy -bench=BenchmarkProxyRoutingLatency -run=^$ -benchmem
```

---

## 5. Coding Standards & Conventions

1. **Zero External Runtime Dependencies:** PORTA compiles into a single static binary. Avoid adding heavy dependencies.
2. **Safe Dialing Invariant:** All outbound HTTP or TCP connections to local services MUST use `security.SafeDialContext()` to prevent SSRF vulnerabilities.
3. **Cardinality Invariant (DECISION SERVICE-001):** Single-service and multi-service workflows must execute through the exact same `registry` and `router` pipelines.

---
*End of Development Guide.*
