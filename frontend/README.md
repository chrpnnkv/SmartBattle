# frontend

React/Vite SPA for SmartBattle. Two main personas:

- **Teacher** — registers, builds quizzes, runs live sessions, views reports.
- **Student** — joins anonymously by PIN, plays the quiz, sees their final rank.

## Stack

- React 18, TypeScript, Vite
- Redux Toolkit (auth/quiz/session slices)
- React Router v6
- CSS Modules
- Vitest + jsdom for tests

## Layout

```
src/
  api/
    IApiService.ts          interface — auth/quizzes/sessions/analytics
    real/realApiService.ts  HTTP impl with 401-redirect handler
    mock/mockApiService.ts  in-memory impl for solo UI work
    wsService.ts            singleton WS client (+ MockWebSocketService)
    index.ts                picks impl by VITE_USE_MOCK
  components/
    layout/                 AppLayout, AuthLayout
    ui/                     Button, Input, Modal, Badge, Logo, ActivityChart
  pages/                    one file per route
  store/
    slices/                 auth, quiz, session
    index.ts                store config
  router/AppRouter.tsx      route table + ProtectedRoute / PublicOnlyRoute
  types/index.ts            shared TS types (mirror server wire shapes)
  test/setup.ts             vitest setup
docs/                       core_integration.md, ws_integration.md
```

## Running

Recommended (real backend in Docker, FE in Vite dev server):

```bash
docker compose up -d postgres quiz-core realtime
cd frontend
cp .env.example .env.local      # then set VITE_USE_MOCK=false
npm install
npm run dev                     # http://localhost:5173
```

Production-style build inside Docker:

```bash
docker compose up -d frontend   # http://localhost:3000 (nginx)
```

## Env vars

| Var | Default in `.env.example` | Notes |
|---|---|---|
| `VITE_API_BASE_URL` | `http://localhost:8080` | quiz-core base URL |
| `VITE_WS_URL` | `ws://localhost:8081` | realtime base URL |
| `VITE_USE_MOCK` | `true` | **Set to `false` in `.env.local`** to talk to the real backend. Vite reads env at startup — restart the dev server after changing this. |

## Tests

```bash
npm run test
```

Files (4):

- `src/__tests__/JoinPage.test.tsx` — student join form rendering + validation.
- `src/__tests__/wsService.test.ts` — `RealWebSocketService` join handshake, send shape, message routing, disconnect cleanup.
- `src/__tests__/sessionSlice.test.ts` — reducers (resetSession, participantJoined, questionStarted/Ended, leaderboard, etc.).
- `src/__tests__/authSlice.test.ts` — token rotation on `changePassword.fulfilled`, logout clears state + localStorage.

## How it talks to the backends

- REST: `realApiService` for everything in `IAuthApi` / `IQuizApi` / `ISessionApi` / `IAnalyticsApi`. See [`docs/core_integration.md`](docs/core_integration.md).
- WebSocket: `wsService` (singleton) for the live game. See [`docs/ws_integration.md`](docs/ws_integration.md).
- 401 recovery is global: `realApiService.handleUnauthorized` clears the token, dispatches a `sb:unauthorized` event (App.tsx → `dispatch(logout())`), and `window.location.replace('/login?reason=expired')`.

## Mock vs real

`VITE_USE_MOCK=true` swaps the API and WS singletons for in-memory mocks that work offline. `mockApiService` keeps fake users/quizzes/sessions in `localStorage`; `MockWebSocketService` simulates events via `localStorage` "signals" so a teacher window and a student window in the **same browser profile** can play together. Cross-profile or incognito sharing requires real mode.
