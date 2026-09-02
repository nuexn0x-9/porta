# Contributing to PORTA

Thank you for your interest in contributing to **PORTA**! We welcome contributions from the open-source community.

---

## 1. Development Workflow

1. **Fork the Repository:** Create your own fork on GitHub.
2. **Clone & Branch:**
   ```bash
   git clone https://github.com/<your-username>/porta.git
   cd porta
   git checkout -b feat/your-feature-name
   ```
3. **Environment Setup:** Ensure Go 1.22+ is installed.
4. **Make Changes:** Write clean, documented Go code with unit tests.
5. **Verify Tests:**
   ```bash
   go test ./... -v
   go vet ./...
   ```
6. **Submit a Pull Request:** Open a PR against the `main` branch.

---

## 2. Commit Message Conventions

We adhere to the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat: add ngrok tunnel provider driver`
- `fix: resolve race condition in health checker`
- `docs: update configuration reference for strip_path`
- `test: add unit test for IPv6 loopback validation`
- `refactor: clean up process management helpers`

---

## 3. Pull Request Guidelines

- All PRs must pass `go test ./...` and `go vet ./...`.
- New features or bug fixes must be accompanied by appropriate unit or integration tests.
- Maintain the invariant that PORTA has zero heavy runtime dependencies and strictly enforces loopback SSRF isolation.

---
*Thank you for helping make PORTA better!*
