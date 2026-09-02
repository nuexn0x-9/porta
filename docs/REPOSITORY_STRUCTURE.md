# PORTA REPOSITORY STRUCTURE

```text
porta/
├── cmd/
│   └── porta/
│       └── main.go                  # Main binary entry point
├── internal/
│   ├── cli/                         # Cobra CLI commands (init, start, status, logs, config, doctor)
│   ├── config/                      # YAML parser, validator, environment expansion, models
│   ├── registry/                    # In-memory target service collection & state tracking
│   ├── router/                      # Longest Prefix Match (LPM) path dispatcher & trie
│   ├── proxy/                       # Ephemeral gateway, ReverseProxy, WebSocket hijacker, error templates
│   ├── security/                    # SSRF guard, dialer hook, Basic Auth & Token Auth gatekeepers
│   ├── health/                      # TCP ping & HTTP GET health probing engine
│   ├── tunnel/                      # TunnelProvider interface
│   │   └── cloudflare/              # Cloudflare Quick Tunnel driver, installer, watchdog
│   ├── ui/                          # Interactive ANSI TUI & structured JSON logger with token redaction
│   └── utils/                       # Windows/POSIX process managers and loopback helpers
├── docs/                            # PRD, SRS, FRM, Guides, Audits, Benchmarks
│   ├── CONFIGURATION_REFERENCE.md
│   ├── DEVELOPMENT.md
│   ├── DOCUMENTATION_AUDIT_REPORT.md
│   ├── FUNCTIONAL_REQUIREMENTS_FINAL.md
│   ├── MVP_RELEASE_AUDIT_REPORT.md
│   ├── PERFORMANCE_VALIDATION_REPORT.md
│   ├── PRODUCT_REQUIREMENTS_FINAL.md
│   ├── RELEASE_CHECKLIST_v1.0.0.md
│   ├── RELEASE_NOTES_v1.0.0.md
│   ├── REPOSITORY_RELEASE_READINESS.md
│   ├── REPOSITORY_STRUCTURE.md
│   ├── SECURITY_AUDIT_REPORT.md
│   ├── SECURITY_MODEL.md
│   ├── SYSTEM_REQUIREMENTS_FINAL.md
│   ├── TROUBLESHOOTING.md
│   └── USER_GUIDE.md
├── examples/                        # Real-world starter templates
│   ├── single-service/              # Single application (e.g. Next.js / Vite)
│   └── full-stack/                  # Multi-service (Frontend + Backend API + Auth)
├── .gitignore
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── go.mod
└── go.sum
```
