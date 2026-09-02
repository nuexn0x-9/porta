# GETTING STARTED WITH PORTA

This quick guide will take you from zero to running your first public environment with **PORTA** in under 2 minutes.

---

## 1. What You Need
- A local web application or API server running on your machine (e.g. `localhost:3000` or `localhost:8000`).
- The `porta` binary installed and accessible in your system `PATH`.

---

## 2. Step-by-Step Guide

### Step 1: Verify System Health
Run `porta doctor` to ensure your workstation is ready:
```bash
porta doctor
```
You should see green checkmarks for OS architecture, loopback interface, internet connectivity, and the Cloudflare tunnel driver.

### Step 2: Initialize Configuration
In the root directory of your project:
```bash
porta init
```
PORTA automatically scans your local machine for active listening web ports (`3000`, `5173`, `8000`, `8080`, etc.) and generates a starter `porta.yaml`.

### Step 3: Launch PORTA
```bash
porta start
```
PORTA will:
1. Bind a local reverse proxy gateway on a dynamic ephemeral port (`127.0.0.1:0`).
2. Establish an encrypted Cloudflare Quick Tunnel to the public internet.
3. Display the interactive Terminal UI (TUI) with your live public HTTPS URL:
   ```text
   Public URL: https://random-generated-slug.trycloudflare.com
   ```

### Step 4: Share & Test
Open the generated HTTPS URL on your phone, share it with teammates for QA, or send it to external webhooks (Stripe, GitHub).

### Step 5: Stop
Press `Ctrl+C` in your terminal to shut down the tunnel and release local resources.

---

## 3. Next Steps
- [Full User Guide & Tutorials](USER_GUIDE.md)
- [Configuration Reference](CONFIGURATION_REFERENCE.md)
- [Security Model](SECURITY_MODEL.md)
