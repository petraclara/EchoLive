# PulseRoom

Real-time event companion — organizers broadcast live updates; attendees follow on their phones without refreshing.

## Stack

| Layer | Tech |
|-------|------|
| Frontend | Next.js 14, Tailwind CSS, TypeScript |
| Backend | Go (chi), REST + WebSockets |
| Database | PostgreSQL 16 |

## Quick start (local)

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Docker](https://www.docker.com/) (for PostgreSQL)

### 1. Start PostgreSQL

```powershell
cd pulseroom
docker compose up -d postgres
```

### 2. Start the API

```powershell
cd apps/api
$env:DATABASE_URL = "postgres://pulseroom:pulseroom@localhost:5432/pulseroom?sslmode=disable"
$env:JWT_SECRET = "dev-secret-change-in-production"
$env:CORS_ORIGIN = "http://localhost:3000"
$env:WEB_APP_URL = "http://localhost:3000"
$env:MIGRATIONS_DIR = "migrations"
go run ./cmd/server
```

API runs at **http://localhost:8080**

### 3. Start the web app

```powershell
cd apps/web
copy .env.local.example .env.local
npm run dev
```

Web app runs at **http://localhost:3000**

## Demo flow

1. Open http://localhost:3000 → **Get started** → register an organizer account.
2. **Create** an event → open the control room.
3. Click **live** to set the event status.
4. Send a live announcement.
5. On your phone (or another browser tab), go to **Join an event** and enter the 6-character code (or scan the QR).
6. The announcement appears instantly on the attendee feed.

## Project structure

```
pulseroom/
├── apps/
│   ├── api/          # Go REST + WebSocket server
│   └── web/          # Next.js frontend
├── docker-compose.yml
└── README.md
```

## API overview

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/auth/register` | Organizer signup |
| POST | `/v1/auth/login` | Organizer login |
| GET | `/v1/events` | List events (auth) |
| POST | `/v1/events` | Create event |
| POST | `/v1/events/:id/announcements` | Send announcement + WS broadcast |
| POST | `/v1/join` | Attendee join by code |
| GET | `/ws/events/:id?token=` | WebSocket room |

## Environment variables

### API (`apps/api`)

| Variable | Default |
|----------|---------|
| `DATABASE_URL` | `postgres://pulseroom:pulseroom@localhost:5432/pulseroom?sslmode=disable` |
| `JWT_SECRET` | (required in production) |
| `CORS_ORIGIN` | `http://localhost:3000` |
| `WEB_APP_URL` | `http://localhost:3000` |
| `PORT` | `8080` |
| `MIGRATIONS_DIR` | `migrations` |

### Web (`apps/web`)

| Variable | Default |
|----------|---------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` |
| `NEXT_PUBLIC_WS_URL` | `ws://localhost:8080` |

## Deployment (suggested)

- **Web**: Vercel — set `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` to your API host.
- **API**: Fly.io or Railway — expose port 8080 with WebSocket support (idle timeout ≥ 60s).
- **DB**: Neon, Supabase, or RDS PostgreSQL.

## Roadmap

- [ ] Live polls & Q&A
- [ ] File uploads (speaker notes PDF)
- [ ] Browser push notifications
- [ ] Redis pub/sub for multi-instance WebSockets
- [ ] Event analytics dashboard

## License

MIT
