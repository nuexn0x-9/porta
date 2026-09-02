# Full-Stack Multi-Service Example

This example demonstrates exposing a multi-service full-stack stack:
- **Frontend:** Next.js / Vite web application on `http://localhost:3000`
- **Backend API:** FastAPI / Express / Go backend API on `http://localhost:8000`
- **Security:** Protected by HTTP Basic Auth (`admin` / `SecretDemo123`)

## Ingress Routing

- `https://<public-url>/` $\rightarrow$ proxied to Frontend on port `3000`
- `https://<public-url>/api/*` $\rightarrow$ proxied to Backend API on port `8000`

## How to Run

1. Start both local services on your workstation:
   - Frontend on port 3000
   - Backend on port 8000
2. Start PORTA in this directory:
   ```bash
   porta start
   ```
3. Authenticate with the credentials configured in `porta.yaml`.
