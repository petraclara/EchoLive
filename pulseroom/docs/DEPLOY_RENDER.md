# Deploy PulseRoom on Render

This guide deploys **PostgreSQL**, the **Go API**, and the **Next.js** app using [Render](https://render.com).

## Option A — Blueprint (recommended)

### 1. Push code to GitHub

```powershell
cd pulseroom
git init
git add .
git commit -m "Initial PulseRoom"
git remote add origin https://github.com/YOUR_USER/pulseroom.git
git push -u origin main
```

### 2. Create services from Blueprint

1. Go to [dashboard.render.com](https://dashboard.render.com)
2. **New** → **Blueprint**
3. Connect your GitHub repo
4. Render reads `render.yaml` and creates:
   - `pulseroom-db` (PostgreSQL)
   - `pulseroom-api` (Docker / Go API)
   - `pulseroom-web` (Next.js)
5. Click **Apply**

First deploy may take 5–10 minutes. Render links services and sets URLs automatically.

### 3. Verify

| Check | URL |
|-------|-----|
| API health | `https://pulseroom-api.onrender.com/health` |
| Web app | `https://pulseroom-web.onrender.com` |

Register an organizer, create an event, set **live**, send an announcement, and join from another device.

---

## Option B — Manual setup

Use this if Blueprint has issues or you want control over names.

### 1. PostgreSQL

1. **New** → **PostgreSQL**
2. Name: `pulseroom-db`, plan **Free**
3. Copy the **Internal Database URL** (for the API on Render)

### 2. API (Web Service)

1. **New** → **Web Service** → connect repo
2. Settings:

| Field | Value |
|-------|--------|
| Name | `pulseroom-api` |
| Root Directory | `apps/api` |
| Runtime | **Docker** |
| Health Check Path | `/health` |

3. Environment variables:

| Key | Value |
|-----|--------|
| `DATABASE_URL` | *(paste Internal Database URL from step 1)* |
| `JWT_SECRET` | *(Generate in Render — long random string)* |
| `MIGRATIONS_DIR` | `migrations` |
| `WEB_APP_URL` | `https://YOUR-WEB-SERVICE.onrender.com` *(set after web deploy)* |
| `CORS_ORIGIN` | same as `WEB_APP_URL` |
| `API_PUBLIC_URL` | `https://YOUR-API-SERVICE.onrender.com` |

4. Deploy and confirm `/health` returns `{"status":"ok"}`.

### 3. Web (Web Service)

1. **New** → **Web Service** → same repo
2. Settings:

| Field | Value |
|-------|--------|
| Name | `pulseroom-web` |
| Root Directory | `apps/web` |
| Runtime | **Node** |
| Build Command | `npm install && npm run build` |
| Start Command | `npm start` |

3. Environment variables:

| Key | Value |
|-----|--------|
| `NODE_VERSION` | `20` |
| `NEXT_PUBLIC_API_URL` | `https://YOUR-API-SERVICE.onrender.com` |

`NEXT_PUBLIC_WS_URL` is optional — the app derives `wss://` from the API URL.

4. Deploy, then **update the API** with the real `WEB_APP_URL` and `CORS_ORIGIN` from this service’s URL.

---

## Important notes

### Free tier

- Services **spin down** after ~15 minutes of inactivity; first request can take 30–60s.
- Free PostgreSQL expires after **90 days** (export data before then).
- WebSockets work on Render web services; reconnect after cold starts.

### HTTPS & WebSockets

- API URL must be `https://…`
- Browsers use `wss://…` for WebSockets (handled automatically in code).

### Custom domains

In each service → **Settings** → **Custom Domains**, then update:

- API: `API_PUBLIC_URL`, `CORS_ORIGIN`
- Web: `WEB_APP_URL`, `NEXT_PUBLIC_API_URL`

Redeploy both services after changing env vars.

### Troubleshooting

| Problem | Fix |
|---------|-----|
| CORS errors | `CORS_ORIGIN` must exactly match the web URL (no trailing slash). |
| WebSocket fails | Confirm API is `https`; check browser console for `wss://` URL. |
| DB connection failed | Use **Internal** database URL on the API service, not External. |
| Next.js can’t reach API | Rebuild web after setting `NEXT_PUBLIC_API_URL` (build-time variable). |

---

## Environment reference

### API

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `JWT_SECRET` | Yes | Signing key for organizer JWTs |
| `PORT` | Auto | Set by Render |
| `WEB_APP_URL` | Yes | Frontend URL for join links / QR |
| `CORS_ORIGIN` | Yes | Usually same as `WEB_APP_URL` |
| `API_PUBLIC_URL` | Yes | Public API URL (`RENDER_EXTERNAL_URL` works) |
| `MIGRATIONS_DIR` | No | Default `migrations` |

### Web

| Variable | Required | Description |
|----------|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Yes | Public API base URL |
| `NEXT_PUBLIC_WS_URL` | No | Optional; defaults from API URL |
