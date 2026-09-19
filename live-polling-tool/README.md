# Live Polling Tool

A production-oriented live polling application built for the GUVI Developer Internship task. It implements React, Go/Gin, MongoDB and Redis as real application components. The core flow is **Create poll → Share link → Audience votes → Live results**, with no browser refresh required.

## Requirements mapped to the brief
- React frontend and Go/Gin backend are separated.
- MongoDB persists users, polls and votes.
- Redis caches the latest result snapshot and publishes vote events per poll.
- WebSocket clients subscribe to Redis-backed events through the Go service.
- Backend validates poll creation and vote requests.
- JWT authentication protects poll management.
- Duplicate anonymous votes are prevented with a server-issued voter cookie and a MongoDB unique `(poll_id, voter_id)` index.
- Docker Compose starts MongoDB, Redis and the Go backend.
- `/health` reports dependency state.

The assignment explicitly requires all four technologies to do real work and calls for genuinely live results rather than refresh-based polling. fileciteturn0file1L9-L16

## Architecture

```text
React browser
   │ REST (JSON)
   ▼
Go + Gin ───────────────► MongoDB (source of truth)
   │
   │ vote event
   ▼
Redis SET + PUB/SUB
   │
   ▼
Go WebSocket handler
   │
   ▼
Connected React browsers
```

### How a vote travels from one browser to another
1. Browser B sends `POST /api/polls/:id/vote` to Go/Gin.
2. Gin validates the poll, option, status and voter identity.
3. MongoDB inserts the vote and atomically increments the selected option.
4. Go reads the authoritative result from MongoDB.
5. Redis stores that result briefly and publishes it on `poll:<pollId>`.
6. Every Go instance with a WebSocket subscriber receives the Redis message.
7. Each connected browser receives the JSON over WebSocket and updates its React state immediately.
8. A reconnect starts with a fresh result read from MongoDB, so clients converge to the persistent source of truth.

Redis is therefore not ornamental: it is the cross-instance event bus and short-lived live-result cache.

## Project structure

```text
live-polling-tool/
├── backend/
│   ├── cmd/server/main.go
│   ├── config/
│   ├── controllers/
│   ├── database/
│   ├── middleware/
│   ├── models/
│   ├── realtime/
│   ├── repositories/
│   ├── routes/
│   ├── services/
│   ├── tests/
│   ├── Dockerfile
│   ├── .env.example
│   └── go.mod
├── frontend/
│   ├── src/api
│   ├── src/components
│   ├── src/context
│   ├── src/pages
│   ├── src/types
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
├── .env.example
├── README.md
└── INTERVIEW_PREPARATION.md
```

## Local setup

### Option A — Docker dependencies

```bash
docker compose up -d mongodb redis
```

Then run backend and frontend in separate terminals.

### Backend

```bash
cd backend
go mod download
# copy .env.example to .env and export/load variables in your shell
set -a; source .env; set +a
go run ./cmd/server
```

Windows PowerShell example:

```powershell
$env:MONGO_URI="mongodb://localhost:27017"
$env:MONGO_DB="live_polling"
$env:REDIS_URL="redis://localhost:6379"
$env:JWT_SECRET="use-a-long-random-secret"
$env:FRONTEND_URL="http://localhost:5173"
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Set `VITE_API_URL=http://localhost:8080` if it is not already configured.

## Environment variables

Backend:
- `PORT` — default `8080`
- `MONGO_URI` — MongoDB connection string
- `MONGO_DB` — database name
- `REDIS_URL` — Redis URL
- `JWT_SECRET` — long random signing secret
- `FRONTEND_URL` — allowed browser origin
- `COOKIE_SECURE` — set `true` when using HTTPS

Frontend:
- `VITE_API_URL` — public backend base URL

Never commit real `.env` files or credentials.

## API contract

| Method | Endpoint | Auth | Purpose |
|---|---|---|---|
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Login and receive JWT |
| GET | `/api/auth/me` | Yes | Validate current token |
| POST | `/api/auth/logout` | No | Client-side logout response |
| GET | `/api/polls` | Yes | List creator's polls |
| POST | `/api/polls` | Yes | Create poll |
| GET | `/api/polls/:id` | No | Public poll |
| GET | `/api/polls/:id/results` | No | Current result |
| POST | `/api/polls/:id/vote` | No | Submit one vote |
| PATCH | `/api/polls/:id/close` | Yes | Close owned poll |
| DELETE | `/api/polls/:id` | Yes | Delete owned poll |
| WS | `/api/polls/:id/live` | No | Live result stream |
| GET | `/health` | No | Service/dependency health |

JSON errors use `{ "error": { "code": "...", "message": "..." } }`.

## Database design

- `users`: email, bcrypt password hash, timestamps. Email has a unique index.
- `polls`: question, embedded options with stable option IDs and vote counters, owner, status, timestamps.
- `votes`: poll ID, option ID, voter ID and timestamp. A unique compound index on `(poll_id, voter_id)` prevents a browser voter from voting twice on the same poll.

The poll document keeps small option metadata and counters together, which makes result reads efficient. Individual vote records remain separate so the system has an audit trail and duplicate-vote constraint without growing a single poll document for every vote.

## Duplicate-vote strategy

Public voters receive a random `voter_id` HttpOnly cookie. The server—not the browser—decides whether that identity has already voted. MongoDB enforces uniqueness. This is deliberately a practical anonymous-voting strategy, not an identity-proofing system: clearing cookies or changing devices can create another anonymous identity. A future stricter system could require verified accounts or signed one-time participation tokens.

## Security
- bcrypt password hashing.
- JWT validation in Gin middleware.
- Owner checks on close/delete.
- Server-side input validation.
- No secrets in source code.
- CORS restricted to `FRONTEND_URL`.
- MongoDB queries use typed BSON filters rather than string concatenation.
- Sensitive credentials are never logged.

## Testing

```bash
cd backend
go test ./...
go build ./...
```

The test package includes result calculation and option-ID sanity checks. For full acceptance testing, use two browser windows against the same public poll and vote alternately; each window should update without refresh.

Frontend:

```bash
cd frontend
npm install
npm run build
```

## Docker

```bash
docker compose up --build
```

This starts MongoDB, Redis and the backend. Run the React frontend separately with `npm run dev`, or deploy it to a static React host.

## Deployment

A practical production arrangement is:
- React/Vite → static frontend host.
- Go service → container/Go host with WebSocket support.
- MongoDB → managed MongoDB service.
- Redis → managed Redis service.

Configure `VITE_API_URL` on the frontend build environment. Configure `MONGO_URI`, `MONGO_DB`, `REDIS_URL`, `JWT_SECRET`, `FRONTEND_URL`, `PORT`, and `COOKIE_SECURE` on the backend. `FRONTEND_URL` must exactly match the deployed browser origin. The backend platform must allow long-lived WebSocket connections and proxy upgrade headers.

After deployment, verify `/health`, signup/login, poll creation, public voting, and the two-browser live update test.

## Design decisions

Correctness is preferred over using Redis as a second database. MongoDB remains the persistent source of truth for votes and counters. Redis holds a short-lived result snapshot and propagates events. If Redis is unavailable at startup, the backend can still serve REST operations, but live WebSocket updates are unavailable until Redis is restored; this is surfaced through `/health` and the WebSocket endpoint rather than silently pretending to be live.

## Troubleshooting
- **MongoDB disconnected:** verify `MONGO_URI` and that MongoDB is running.
- **Redis disconnected:** verify `REDIS_URL`; live updates require Redis.
- **CORS error:** make `FRONTEND_URL` equal to the exact frontend origin.
- **WebSocket reconnecting:** verify the backend is reachable and the deployment supports WebSockets.
- **401:** sign in again or clear the stale token from local storage.

## Interview reminder
The original brief explicitly notes that AI assistance is acceptable but expects candidates to understand the fundamentals of what they build, and it requires the submission video. fileciteturn0file1L54-L60
