# WebSocket Protocol

`realtime` exposes a WebSocket endpoint at `ws://<host>/ws`. The first frame the client sends MUST be `join`; all other frames are rejected until then.

## Handshake — `join`

```json
{
  "type": "join",
  "room_code": "ABCD12",
  "name": "Алиса",
  "token": "<JWT>",
  "participant_id": "<UUID from POST /api/sessions/join>"
}
```

| Field | Sender | Notes |
|---|---|---|
| `room_code` | both | PIN of the room. Whitespace and dashes ignored, uppercased server-side. |
| `name` | student (required), teacher (optional) | Display name. Teacher: if missing AND JWT carries `email`, server uses email. Otherwise falls back to `Преподаватель`. |
| `token` | teacher | JWT issued by `quiz-core`. Without it the connection is treated as a student. |
| `participant_id` | student | Stable UUID returned by `POST /api/sessions/join`. If absent, server generates an ephemeral one. |

Server response on success:

```json
{
  "type": "joined",
  "timestamp": "...",
  "payload": {
    "room_code": "ABCD12",
    "role": "student",
    "name": "Алиса",
    "quiz_title": "История",
    "total_questions": 5,
    "participants": [{ "id": "uuid", "nickname": "Боб", "avatarInitials": "БО",
                       "avatarColor": "#7c3aed", "score": 0, "answeredCount": 0 }],
    "totalCount": 1
  }
}
```

`participants` and `totalCount` reflect the **current** student roster at the moment the new client joined — used by both teacher (to seed the count after a reconnect) and students (to know who's already in the lobby).

Server response on failure:

```json
{ "type": "error", "payload": { "code": "room_not_found", "message": "комната не найдена: ABCD12" } }
```

then a normal close frame. Codes: `invalid_token`, `room_not_found`, `room_full`, `session_not_active`, `invalid_message`.

## Client → Server

| Type | Body | Sender | Effect |
|---|---|---|---|
| `answer` | `{ question_id, answer_id }` (or `answer_index`) | student | Records answer, sends `answer_result` back to student and `answer_received` to teacher. |
| `start_session` | none | teacher | Starts the quiz and broadcasts `session_started` + first `question_started`. |
| `next_question` | none | teacher | Closes current question (`question_ended`), opens next (or finishes). |
| `finish_session` | none | teacher | Force-finishes the session (broadcasts `question_ended` for the current question first). |
| `ping` | none | both | Heartbeat; server replies with `pong`. |

## Server → Client

All frames share the envelope `{ type, timestamp, payload }`. Payload shapes:

### `participant_joined`
```json
{ "participant": { "id", "nickname", "avatarInitials", "avatarColor", "score", "answeredCount" },
  "totalCount": 3 }
```

### `participant_left`
```json
{ "participant_id": "uuid", "name": "Алиса", "totalCount": 2 }
```

### `session_started`
```json
{ "quiz_title": "История", "total_questions": 5 }
```

### `question_started`
```json
{
  "question": {
    "id": "q1", "quizId": "quiz-1", "type": "multiple_choice",
    "text": "Сколько будет 2+2?",
    "imageUrl": "https://...",
    "options": [
      { "id": "q1_o0", "text": "3", "isCorrect": false, "color": "red" },
      { "id": "q1_o1", "text": "4", "isCorrect": true,  "color": "blue" }
    ],
    "timeLimitSeconds": 30,
    "order": 0
  },
  "questionIndex": 0,
  "totalQuestions": 5,
  "startedAt": 1713100000000
}
```
`questionIndex` is 0-based; `startedAt` is epoch ms.

### `answer_received` (teacher only)
```json
{ "participant_name": "Алиса", "answers_count": 3, "total_participants": 5 }
```

### `answer_result` (student only)
```json
{ "correct": true, "correct_index": 1, "score": 850, "total_score": 850 }
```

### `question_ended`
```json
{
  "questionReport": {
    "questionId": "q1", "questionText": "Сколько будет 2+2?",
    "correctPercent": 80, "avgResponseTimeMs": 4200,
    "mostCommonWrongOptionId": "q1_o0", "mostCommonWrongOptionText": "3",
    "distribution": [
      { "optionId": "q1_o0", "optionText": "3", "count": 1, "isCorrect": false, "color": "red" },
      { "optionId": "q1_o1", "optionText": "4", "count": 4, "isCorrect": true,  "color": "blue" }
    ],
    "fastestCorrectParticipants": [{ "id": "uuid", "nickname": "Боб" }]
  },
  "leaderboard": [
    { "id": "uuid", "nickname": "Боб", "avatarInitials": "БО", "avatarColor": "#2563eb",
      "score": 950, "answeredCount": 1 }
  ],
  "endedAt": 1713100030000
}
```

### `leaderboard`
```json
{ "entries": [{ "rank": 1, "name": "Боб", "score": 950 }] }
```
Broadcast after each `question_ended`.

### `session_finished`
```json
{
  "quiz_title": "История",
  "duration_sec": 180,
  "results": [{ "name": "Боб", "score": 4200, "correct_answers": 4, "total_questions": 5 }]
}
```
Same `results` array is also POSTed to `quiz-core` `/internal/reports`, where it ends up in `GameReport.leaderboard`.

### `error`
```json
{ "code": "rate_limit_exceeded", "message": "слишком много сообщений" }
```
Server-side errors from any of the validation layers. Followed by a normal close frame so the client's `onclose` fires with `wasClean=true`.

### `pong`
No payload.

## REST companion endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/api/rooms` | Bearer JWT (teacher) | `quiz-core` calls this to create a room. Returns `{ room_code, host_id, quiz_mode, status }`. |
| GET  | `/api/rooms` | none | Diagnostic: returns `{ active_rooms: N }`. |
| GET  | `/api/rooms/:code` | none | Diagnostic: room status, participant count, host_id. |
| GET  | `/api/rooms/:code/participants` | none | Live participants + `current_question_index`. Used by `quiz-core` `BuildSessionDTO`. |
| GET  | `/health` | none | `{ status: "ok", active_rooms, timestamp }`. |