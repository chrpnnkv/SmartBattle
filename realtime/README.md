# SmartBattle — Backend Real-Time Module

Backend-модуль реального времени для веб-платформы проведения академических квизов «Smart Battle».

Реализован на **Go 1.22** с использованием протокола **WebSocket** (gorilla/websocket).

---

## Архитектура

```
HTTP :8080                                                
├── POST /api/rooms      — создание игровой комнаты       
├── GET  /api/rooms/{code} — информация о комнате         
├── GET  /health          — healthcheck                   
└── GET  /ws              — WebSocket endpoint            
                                                          
┌──────────┐  ┌──────────┐  ┌──────────┐                 
│  Client  │  │   Room   │  │ RoomMgr  │                 
│ (WS conn)│  │(State    │  │(registry)│                 
│ readPump │  │ Machine) │  │          │                 
│ writePump│  │Waiting → │  │          │                 
└──────────┘  │Active →  │  └──────────┘                 
              │Finished  │                                 
              └──────────┘                                 
                    │                                      
                    ▼                                      
           ┌─────────────────┐                            
           │  backend-core   │  POST /api/internal/       
           │  HTTP client    │       quiz-results         
           └─────────────────┘                            

```

### Структура проекта

```
smartbattle-realtime/
├── cmd/server/main.go              # Точка входа, graceful shutdown
├── internal/
│   ├── config/config.go            # Конфигурация из env
│   ├── message/types.go            # Протокол WebSocket-сообщений
│   ├── auth/jwt.go                 # Проверка JWT-токенов
│   ├── client/client.go            # WS-клиент (read/write pumps)
│   ├── room/
│   │   ├── room.go                 # State machine игровой комнаты
│   │   ├── manager.go              # Реестр комнат (потокобезопасный)
│   │   ├── broadcast.go            # Рассылка сообщений участникам
│   │   └── export.go               # Публичные методы для core-клиента
│   ├── handler/handler.go          # HTTP/WS обработчики
│   └── core/client.go              # HTTP-клиент для backend-core
├── pkg/ratelimit/limiter.go        # Rate limiter (token bucket)
├── internal/auth/jwt_test.go       # Тесты JWT
├── internal/room/room_test.go      # Тесты Room и Manager
├── pkg/ratelimit/limiter_test.go   # Тесты Rate Limiter
├── .env.example
├── Dockerfile
└── docker-compose.yml
```

---

## Протокол WebSocket

### Входящие сообщения (Client → Server)

#### Вход в комнату (обязательно первым сообщением)
```json
// Студент
{"type": "join", "room_code": "ABCD12", "name": "Иван Петров"}

// Преподаватель
{"type": "join", "room_code": "ABCD12", "token": "<JWT>"}
```

#### Управление квизом (только преподаватель)
```json
{"type": "start_session"}
{"type": "next_question"}
{"type": "finish_session"}
```

#### Ответ студента
```json
{"type": "answer", "question_id": "<uuid>", "answer_index": 2}
```

#### Heartbeat
```json
{"type": "ping"}
```

---

### Исходящие сообщения (Server → Client)

#### Подтверждение входа
```json
{
  "type": "joined",
  "timestamp": "...",
  "payload": {
    "room_code": "ABCD12",
    "role": "student",
    "name": "Иван",
    "quiz_title": "Математика",
    "total_questions": 10
  }
}
```

#### Новый вопрос
```json
{
  "type": "question",
  "payload": {
    "question_id": "uuid",
    "index": 1,
    "total": 10,
    "text": "Сколько будет 2+2?",
    "options": [{"index": 0, "text": "3"}, {"index": 1, "text": "4"}],
    "time_limit_sec": 30,
    "started_at": "..."
  }
}
```

#### Результат ответа (студенту)
```json
{
  "type": "answer_result",
  "payload": {
    "correct": true,
    "correct_index": 1,
    "score": 950,
    "total_score": 1900
  }
}
```

#### Итоги вопроса (всем)
```json
{
  "type": "question_results",
  "payload": {
    "question_id": "uuid",
    "correct_index": 1,
    "stats": [{"option_index": 0, "count": 3}, {"option_index": 1, "count": 15}],
    "leaderboard": [{"rank": 1, "name": "Иван", "score": 950}]
  }
}
```

#### Завершение сессии (всем)
```json
{
  "type": "session_finished",
  "payload": {
    "quiz_title": "Математика",
    "duration_sec": 180,
    "results": [
      {"name": "Иван", "score": 4200, "correct_answers": 5, "total_questions": 5}
    ]
  }
}
```

---

## Старт

### 1. Установка зависимостей

```bash
go mod tidy
```

### 2. Конфигурация

```bash
cp .env.example .env
# Обязательно задайте JWT_SECRET
```

### 3. Запуск

```bash
JWT_SECRET=your_secret go run ./cmd/server
```

### 4. Запуск через Docker

```bash
docker build -t smartbattle-realtime .
docker run -p 8080:8080 -e JWT_SECRET=your_secret smartbattle-realtime
```

### 5. Docker Compose (с заглушкой backend-core)

```bash
JWT_SECRET=your_secret docker compose --profile dev up
```

---

## Сценарий использования

```
1. Преподаватель создаёт комнату:
   POST /api/rooms
   Authorization: Bearer <teacher_jwt>
   Body: {"quiz_id":"...", "quiz_title":"...", "questions":[...]}
   → {"room_code": "ABCD12", ...}

2. Преподаватель подключается по WS:
   ws://host/ws
   → {"type":"join","room_code":"ABCD12","token":"<jwt>"}

3. Студенты подключаются по WS:
   → {"type":"join","room_code":"ABCD12","name":"Иван"}

4. Преподаватель запускает:
   → {"type":"start_session"}

5. Студенты отвечают:
   → {"type":"answer","question_id":"...","answer_index":1}

6. Преподаватель переключает вопросы:
   → {"type":"next_question"}

7. Преподаватель завершает:
   → {"type":"finish_session"}
   ← Всем: session_finished + результаты → backend-core
```

---

## Тестирование

```bash
# Все тесты
go test ./...

# С verbose выводом
go test -v ./...

# Конкретный пакет
go test -v ./internal/auth/...
go test -v ./internal/room/...
go test -v ./pkg/ratelimit/...

# С покрытием
go test -cover ./...
```

---

## Переменные окружения

| Переменная              | По умолчанию        | Описание                             |
|-------------------------|---------------------|--------------------------------------|
| `JWT_SECRET`            | **обязательно**     | Секрет для проверки JWT              |
| `PORT`                  | `8080`              | Порт сервера                         |
| `BACKEND_CORE_URL`      | `http://localhost:8081` | URL backend-core                 |
| `MAX_PARTICIPANTS`      | `100`               | Макс. студентов в комнате            |
| `DEFAULT_QUESTION_TIME_SEC` | `30`           | Время на ответ по умолчанию (сек)    |
| `RATE_LIMIT_MESSAGES`   | `10`                | Макс. сообщений в период             |
| `RATE_LIMIT_PERIOD`     | `1s`                | Период rate limit                    |
| `LOG_LEVEL`             | `info`              | Уровень логирования (debug/info)     |

