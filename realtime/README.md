# realtime

WebSocket gateway for live quiz rooms. Stateless (rooms are in-memory only); persistent state lives in `quiz-core`.

## Stack

- Go 1.22, gorilla/websocket, log/slog
- No DB. Rooms are kept in `room.Manager` (a `map[string]*Room` behind a mutex).

## Layout

```
cmd/server/           entrypoint
internal/
  auth/               JWT verify (HS256, shared secret with quiz-core)
  client/             WS client wrapper (read/write loops, channel-backed SendMsg)
  config/             env loader
  core/               HTTP client to quiz-core (POST /internal/reports etc.)
  handler/            WS upgrade + REST /api/rooms*
  message/            wire types (join/answer/question_started/...)
  room/               Room (state machine, scoring, broadcast) + Manager
pkg/ratelimit/        token-bucket per WS client
docs/ws_events.md     wire-protocol reference
```

## Running

Standalone:

```bash
cd realtime
cat > .env <<EOF
PORT=8081
HOST=0.0.0.0
JWT_SECRET=supersecret
BACKEND_CORE_URL=http://localhost:8080
BACKEND_CORE_INTERNAL_SECRET=internal_secret_key
EOF
go mod download
go run ./cmd/server
```

Inside Docker (recommended): `docker compose up -d realtime` from repo root.

## Env vars

| Var | Default | Notes |
|---|---|---|
| `HOST` | `0.0.0.0` | bind address |
| `PORT` | `8080` | bind port (mapped to host 8081 in compose) |
| `JWT_SECRET` | required | must match quiz-core |
| `BACKEND_CORE_URL` | `http://localhost:8081` (sic — set explicitly!) | base URL of quiz-core |
| `BACKEND_CORE_INTERNAL_SECRET` | — | header value for `X-Internal-Secret` |
| `BACKEND_CORE_TIMEOUT` | `5s` | HTTP timeout for outbound calls |
| `MAX_PARTICIPANTS` | `100` | per-room cap |
| `RATE_LIMIT_MESSAGES` / `RATE_LIMIT_PERIOD` | `10` / `1s` | per-client WS frame limit |
| `DEFAULT_QUESTION_TIME_SEC` | `30` | fallback when a question has no `timeLimitSeconds` |
| `ROOM_CODE_LENGTH` | `6` | digits in PINs |
| `LOG_LEVEL` | `info` | `debug` enables verbose slog |

## Endpoints

WebSocket: `GET /ws` (also `/ws/...`). First frame must be `join`. Full protocol in [`docs/ws_events.md`](docs/ws_events.md).

REST:
- `POST /api/rooms` — Bearer JWT (teacher). Used by quiz-core to create a room. Returns `{ room_code, quiz_id, quiz_title, quiz_mode, host_id, status, ws_url }`.
- `GET /api/rooms` — `{ active_rooms: N }` (diagnostic).
- `GET /api/rooms/:code` — room status snapshot.
- `GET /api/rooms/:code/participants` — `{ participants: [...], current_question_index }`. quiz-core uses this in `BuildSessionDTO`.
- `GET /health` — `{ status: "ok", active_rooms, timestamp }`.

## Tests

```bash
go test ./...
```

- `internal/auth/jwt_test.go` — token verify happy/expired/wrong-secret.
- `internal/room/room_test.go` — initial state, manager CRUD.
- `internal/room/score_test.go` — `calcScore` formula across timing edge cases.
- `pkg/ratelimit/limiter_test.go` — token bucket allow/refill.

## Notable design choices

- `Room` carries one `sync.RWMutex`. All mutating methods are split into `*` (acquires lock) + `*Locked` (assumes lock held). Anything that touches `Room.Participants`, `Status`, `CurrentQuestionIndex` or `QuestionReports` must run inside the write lock.
- `client.SendMsg` is non-blocking — it pushes to a buffered channel; the per-client write goroutine drains it. This lets `broadcastLocked` send to N participants while still holding the room lock.
- Per-question reports are accumulated in `Room.QuestionReports` and shipped to quiz-core in the finish payload, populating `GameReport.questionReports` on the FE.
- Session results are POSTed to quiz-core with **3 attempts and exponential backoff** (1s/2s/4s) — see `core.SaveResultsPayloadWithRetry`.
- WS upgrade requires the response writer to satisfy `http.Hijacker`. The logging middleware in `cmd/server/main.go` wraps it but explicitly forwards `Hijack()` and `Flush()` — without that the `gorilla/websocket` upgrade fails with `response does not implement http.Hijacker`.
- Healthcheck in `docker-compose.yml` uses `wget` from busybox; the Dockerfile uses `alpine:3.20` rather than `scratch` so the healthcheck command is available.
