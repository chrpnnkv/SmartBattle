# WebSocket — интеграция фронтенда с реалтаймом

Документ описывает найденные расхождения между фронтендом и реалтаймом и что было исправлено.
Основан на `realtime/docs/ws_events.md`.

---

## Найденные проблемы

### 1. Отсутствовал `join`-хендшейк (критично)

`RealWebSocketService.connect()` открывал соединение, но не отправлял первое сообщение `join`.
Сервер ожидает его **сразу после открытия** соединения:

```json
{ "type": "join", "room_code": "ABCD12", "name": "Алиса", "token": "<JWT>" }
```

**Исправлено:** в `onopen` теперь отправляется `join` с `room_code`, `name`, `token`.

---

### 2. Неправильный формат исходящих сообщений (критично)

Фронтенд отправлял:
```json
{ "type": "answer", "payload": { "question_id": "q1", "answer_id": "q1_o2" } }
```

Сервер ожидает плоский формат:
```json
{ "type": "answer", "question_id": "q1", "answer_id": "q1_o2" }
```

**Исправлено:** `send()` теперь делает `{ type, ...payload }` вместо `{ type, payload }`.

---

### 3. PIN не сохранялся для WS-подключения студента

После `joinSession` (REST) студент переходил на страницу ожидания, но PIN нигде не сохранялся.
При WS-подключении нужно передать `room_code` (PIN) в `join`.

**Исправлено:**
- `JoinPage` сохраняет PIN в `sessionStorage.setItem('sb_pin', ...)` при сабмите
- `WaitingRoomPage` и `QuestionPage` читают его при вызове `wsService.connect()`

---

### 4. Учитель использовал мок-события вместо реальных (важно)

`AnalyticsPage` отправлял события `start_question` и `end_question` — они существуют только
в `MockWebSocketService` и серверу неизвестны.

| Было (мок) | Стало (реальный сервер) |
|---|---|
| `wsService.send('start_question', { question, ... })` | `wsService.send('start_session', {})` + `wsService.send('next_question', {})` |
| `wsService.send('end_question', ...)` при конце таймера | `wsService.send('next_question', {})` или `wsService.send('finish_session', {})` |
| `dispatch(endSession(id))` REST | `wsService.send('finish_session', {})` |

**Исправлено:** в `AnalyticsPage` через флаг `USE_MOCK` выбирается нужный путь.
Мок-режим (`VITE_USE_MOCK=true`) работает без изменений.

---

### 5. Ответы студентов не шли через WS

Ответы отправлялись только через REST `api.sessions.submitAnswer()`.
Сервер ожидает ответ по WS:

```json
{ "type": "answer", "question_id": "q1", "answer_id": "q1_o2" }
```

**Исправлено:** в `QuestionPage` перед REST-вызовом добавлен:
```ts
wsService.send('answer', { question_id: currentQuestion.id, answer_id: optionId });
```
REST-вызов оставлен как фолбэк для мок-режима.

---

### 6. Не обрабатывались события от сервера

| Событие | Где добавлено | Что делает |
|---|---|---|
| `answer_result` | `QuestionPage` | Обновляет правильность и очки из ответа сервера, перезаписывая клиентский расчёт |
| `answer_received` | `AnalyticsPage` (real-режим) | Обновляет счётчик ответивших; в student_paced автоматически завершает вопрос |
| `question_started` | `AnalyticsPage` (real-режим) | Синхронизирует индекс вопроса и запускает таймер учителя по событию сервера |

---

### 7. TS-ошибка в sessionSlice

`state.participants = []` — поля `participants` нет в `SessionState` (участники хранятся в `state.session.participants`).
**Исправлено:** строки удалены.

---

## Проверено по исходникам realtime (room.go, handler.go, types.go)

- `start_session` **автоматически запускает первый вопрос** — `StartSession()` сразу вызывает `nextQuestionLocked()`. Отдельный `next_question` для Q1 не нужен и был бы ошибкой (перешёл бы на Q2).
- `next_question` на **последнем вопросе** — сервер сам вызывает `finishLocked()`. Явный `finish_session` нужен только для досрочного завершения.
- `name` в `join` **обязателен**. Для учителя: если `name=""` и есть JWT — берёт `email` из claims. Невалидный токен → соединение закрывается с `error { code: "invalid_token" }`.
- `joined` присылается сразу после успешного join, до любых других событий.
- REST-ответ (`SubmitAnswerREST`) тоже отправляет `answer_result` по WS — дублирования не будет, придёт ровно одно событие.

### Открытые вопросы и зависимости не от фронтенда

#### Критично — интеграция двух бэкендов

Когда учитель вызывает `POST /api/sessions` на основном бэке (порт 8080), тот **должен создать комнату** в realtime-сервере (`POST /api/rooms`, порт 8081) с **тем же кодом**, что вернул в `session.pin`. Если этого нет — фронт пошлёт `join` с несуществующим `room_code` и получит `error { code: "room_not_found" }`.

#### Критично — общий JWT-секрет

Realtime проверяет токен через свой `auth.Service`. Если JWT-секрет отличается от основного бэка — учитель получит `error { code: "invalid_token" }` при подключении.

#### Важно — PIN у студента после перезагрузки страницы

Студент получает `room_code` (PIN) из `sessionStorage` (`sb_pin`), который сохраняется при вводе на JoinPage. Если студент перезагрузит WaitingRoomPage или QuestionPage — `sb_pin` пропадёт и WS-подключение упадёт с `room_not_found`. Для полного решения нужно либо хранить PIN в Redux/localStorage, либо возвращать его из REST `/api/sessions/join`.

#### Для коллеги (realtime)

1. **Двойное WS-подключение студента** — при переходе `WaitingRoomPage` → `QuestionPage` происходит `disconnect()` + новый `connect()` с повторным `join`. По коду (`OnDisconnect` удаляет старый `client.ID`, новый добавляется) — должно работать, но стоит проверить на практике.

2. **`leaderboard` событие** — в `types.go` есть `TypeLeaderboard = "leaderboard"`, но в `room.go` оно нигде не отправляется. Планируется? На фронтенде не обрабатывается.

---

## Изменённые файлы

| Файл | Что изменено |
|---|---|
| `src/api/IApiService.ts` | Добавлен `WsConnectOptions`, обновлена сигнатура `connect()` |
| `src/api/wsService.ts` | `RealWebSocketService`: join-хендшейк, плоский send, ping; экспорт `USE_MOCK` |
| `src/store/slices/sessionSlice.ts` | Убраны `state.participants = []` |
| `src/pages/JoinPage.tsx` | Сохранение PIN в `sessionStorage` |
| `src/pages/WaitingRoomPage.tsx` | `connect()` с `roomCode`/`name`; обработка `error`, `joined` |
| `src/pages/QuestionPage.tsx` | `connect()` с опциями; WS-отправка ответа; обработка `answer_result`, `error`, `joined` |
| `src/pages/AnalyticsPage.tsx` | `connect()` с `roomCode`/`token`; `start_session`/`next_question`/`finish_session`; `answer_received`; `question_started`; обработка `error`, `joined` |

---

## Переменные окружения

```env
VITE_USE_MOCK=false          # переключить на реальный бэкенд
VITE_API_URL=http://host:port
VITE_WS_URL=ws://host:port
```
