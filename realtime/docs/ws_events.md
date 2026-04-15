# WebSocket Protocol

## Подключение

Клиент подключается к `ws://<host>/ws` (или `/ws/<sessionId>` — путь игнорируется).

**Первым сообщением** клиент обязан отправить `join`:

```json
{
  "type": "join",
  "room_code": "ABCD12",
  "name": "Алиса",
  "token": "<JWT>"
}
```

| Поле | Кто передаёт | Описание |
|---|---|---|
| `room_code` | все | PIN/код комнаты (полученный при создании сессии) |
| `name` | студент | никнейм |
| `token` | преподаватель | JWT-токен авторизации |

---

## Входящие сообщения (клиент → сервер)

### `join` — вход в комнату
```json
{ "type": "join", "room_code": "ABCD12", "name": "Алиса", "token": "" }
```

### `answer` — ответ студента
```json
{ "type": "answer", "question_id": "q1", "answer_id": "q1_o2" }
```
> `answer_id` — ID варианта ответа из payload вопроса. Можно передать `answer_index` (int, 0-based) как запасной вариант.

### `start_session` — запустить квиз (учитель)
```json
{ "type": "start_session" }
```

### `next_question` — следующий вопрос (учитель)
```json
{ "type": "next_question" }
```

### `finish_session` — завершить сессию (учитель)
```json
{ "type": "finish_session" }
```

### `ping` — heartbeat
```json
{ "type": "ping" }
```

---

## Исходящие сообщения (сервер → клиент)

Все сообщения имеют общую обёртку:
```json
{ "type": "...", "timestamp": "2026-04-14T...", "payload": { ... } }
```

---

### `joined` — подтверждение входа
```json
{
  "room_code": "ABCD12",
  "role": "student",
  "name": "Алиса",
  "quiz_title": "История",
  "total_questions": 5
}
```

---

### `participant_joined` — новый участник подключился
Совместим с фронтендом `WsParticipantJoinedPayload`.
```json
{
  "participant": {
    "id": "uuid",
    "nickname": "Алиса",
    "avatarInitials": "АЛ",
    "avatarColor": "#7c3aed",
    "score": 0,
    "answeredCount": 0
  },
  "totalCount": 3
}
```

---

### `session_started` — сессия запущена
```json
{ "quiz_title": "История", "total_questions": 5 }
```

---

### `question_started` — новый вопрос
Совместим с фронтендом `WsQuestionStartedPayload`.
```json
{
  "question": {
    "id": "q1",
    "quizId": "quiz-1",
    "type": "multiple_choice",
    "text": "Сколько будет 2+2?",
    "options": [
      { "id": "q1_o0", "text": "3", "isCorrect": false, "color": "red" },
      { "id": "q1_o1", "text": "4", "isCorrect": true,  "color": "blue" },
      { "id": "q1_o2", "text": "5", "isCorrect": false, "color": "yellow" },
      { "id": "q1_o3", "text": "6", "isCorrect": false, "color": "green" }
    ],
    "timeLimitSeconds": 30,
    "order": 0
  },
  "questionIndex": 0,
  "totalQuestions": 5,
  "startedAt": 1713100000000
}
```
> `questionIndex` — 0-based. `startedAt` — epoch ms.

---

### `answer_result` — результат ответа (студенту)
```json
{ "correct": true, "correct_index": 1, "score": 850, "total_score": 850 }
```

---

### `answer_received` — ответ поступил (учителю)
```json
{ "participant_name": "Алиса", "answers_count": 3, "total_participants": 5 }
```

---

### `question_ended` — итоги вопроса
Совместим с фронтендом `WsQuestionEndedPayload`.
```json
{
  "questionReport": {
    "questionId": "q1",
    "questionText": "Сколько будет 2+2?",
    "correctPercent": 80,
    "avgResponseTimeMs": 4200,
    "mostCommonWrongOptionId": "q1_o0",
    "mostCommonWrongOptionText": "3",
    "distribution": [
      { "optionId": "q1_o0", "optionText": "3", "count": 1, "isCorrect": false, "color": "red" },
      { "optionId": "q1_o1", "optionText": "4", "count": 4, "isCorrect": true,  "color": "blue" }
    ],
    "fastestCorrectParticipants": [
      { "id": "uuid", "nickname": "Боб" }
    ]
  },
  "leaderboard": [
    { "id": "uuid", "nickname": "Боб", "avatarInitials": "БО", "avatarColor": "#2563eb", "score": 950, "answeredCount": 1 }
  ],
  "endedAt": 1713100030000
}
```

---

### `session_finished` — сессия завершена
```json
{
  "quiz_title": "История",
  "results": [
    { "name": "Боб", "score": 4200, "correct_answers": 4, "total_questions": 5 }
  ],
  "duration_sec": 180
}
```

---

### `error` — ошибка
```json
{ "code": "room_not_found", "message": "комната не найдена" }
```

### `pong` — ответ на heartbeat
Нет payload.

---

## Что нужно изменить на фронтенде (для коллеги)

### `wsService.ts` — `RealWebSocketService`

1. **`connect(sessionId, participantId?)`** — после открытия соединения отправить join:
   ```ts
   this.socket.onopen = () => {
     this.socket.send(JSON.stringify({
       type: 'join',
       room_code: this.roomCode, // PIN комнаты, нужно передавать в connect()
       name: this.nickname,      // никнейм студента
       token: this.token,        // JWT для преподавателя, '' для студента
     }));
   };
   ```
   Сигнатуру лучше изменить на `connect(roomCode: string, options: { name?: string; token?: string })`.

2. **`send('answer', ...)`** — уже используется, изменить payload:
   ```ts
   // Вместо POST /api/sessions/answer:
   wsService.send('answer', { question_id: questionId, answer_id: optionId });
   ```

3. **Учитель** — заменить REST-вызовы на WS:
   - `api.sessions.startSession(id)` → `wsService.send('start_session', {})`
   - `api.sessions.nextQuestion(id)` → `wsService.send('next_question', {})`
   - `api.sessions.endSession(id)` → `wsService.send('finish_session', {})`

4. **Событие `question_started`** — уже подписано в `WaitingRoomPage` и `QuestionPage`. Payload теперь совместим.

5. **Событие `question_ended`** — уже подписано в `QuestionPage`. Payload теперь совместим.
