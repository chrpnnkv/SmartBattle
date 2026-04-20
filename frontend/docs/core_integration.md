# REST — интеграция фронтенда с backend-core

Документ описывает расхождения между фронтендом и quiz-core (порт 8080).
Основан на `quiz-core/internal/transport/rest/handler/handler.go` и `quiz-core/internal/models/models.go`.

---

## Модели: что думает фронт vs что отдаёт бэк

### Quiz

| Поле (фронт) | Поле (бэк) | Проблема |
|---|---|---|
| `id: string` (UUID) | `id: int64` | **Критично** — типы не совпадают |
| `title: string` | `title: string` | ✓ |
| `description?: string` | `description: string` | ✓ |
| `settings: QuizSettings` | `settings: datatypes.JSON` | Бэк хранит как raw JSON, структура на его усмотрение |
| `questions: Question[]` | `questions: []Question` | Структуры вопросов тоже расходятся (см. ниже) |
| `status: 'draft'/'published'` | **отсутствует** | Бэк не знает про статус |
| `mode: 'teacher_paced'/'student_paced'` | **отсутствует** | Бэк не знает про режим |
| `questionCount: number` | **отсутствует** | Вычисляется на фронте |
| `estimatedMinutes: number` | **отсутствует** | Вычисляется на фронте |
| `authorId: string` | `teacher_id: uuid.UUID` | Разные имена |
| `createdAt: string` | `created_at: time.Time` | Разные имена (snake_case vs camelCase) |
| `updatedAt: string` | **отсутствует** | Бэк не хранит |

### Question

| Поле (фронт) | Поле (бэк) | Проблема |
|---|---|---|
| `id: string` | `id: uuid` | ✓ |
| `quizId: string` | `quiz_id: int64` | Тип `quizId` должен быть числом |
| `type: QuestionType` | `type: string` | ✓ |
| `text: string` | `text: string` | ✓ |
| `timeLimitSeconds: number` | `timer_sec: int` | **Критично** — разные имена |
| `order: number` | `order: int` | ✓ |
| `options: AnswerOption[]` | `options: []Option` | Структуры расходятся (см. ниже) |
| `imageUrl?: string` | **отсутствует** | Бэк не хранит |

### AnswerOption

| Поле (фронт) | Поле (бэк) | Проблема |
|---|---|---|
| `id: string` | `id: string` | ✓ |
| `text: string` | `text: string` | ✓ |
| `isCorrect: boolean` | `is_correct: bool` | Разные имена |
| `color: 'red'/'blue'/'yellow'/'green'` | **отсутствует** | Бэк не хранит цвет |

---

## Эндпоинты: что есть на бэке, что ожидает фронт

### Auth

| Метод | URL (бэк) | URL (фронт) | Статус |
|---|---|---|---|
| POST | `/auth/register` | `/auth/register` | ✓ |
| POST | `/auth/login` | `/auth/login` | ✓ — но бэк возвращает `{ token }`, фронт ожидает `{ user, tokens: { accessToken } }` |
| GET | `/api/me` | `/api/me` | ✓ |
| POST | `/auth/change-password` | `/auth/change-password` | ✓ |
| POST | `/auth/forgot-password` | `/auth/forgot-password` | ✓ |
| POST | `/auth/reset-password` | `/auth/reset-password` | ✓ |

### Quizzes

| Метод | URL (бэк) | URL (фронт) | Статус |
|---|---|---|---|
| POST | `/api/quizzes` | `/api/quizzes` | ✓ URL совпадает, но модель расходится |
| PUT | `/api/quizzes/{id}` | `/api/quizzes/${id}` | ✓ |
| GET | `/api/quizzes` | `/api/quizzes` | ✓ |
| GET | `/api/quizzes/public` | `/api/quizzes/public` | ✓ |
| GET | `/api/quizzes/{id}` | `/api/quizzes/${id}` | ✓ |
| DELETE | `/api/quizzes/{id}` | `/api/quizzes/${id}` | ✓ |

### Sessions

| Метод | URL (бэк) | URL (фронт) | Статус |
|---|---|---|---|
| POST | **отсутствует** | `/api/sessions` | **Критично** — бэк не реализовал сессии |
| POST | **отсутствует** | `/api/sessions/join` | **Критично** |
| GET | **отсутствует** | `/api/sessions/${id}` | **Критично** |
| POST | **отсутствует** | `/api/sessions/${id}/start` | **Критично** |
| POST | **отсутствует** | `/api/sessions/${id}/next` | **Критично** |
| POST | **отсутствует** | `/api/sessions/${id}/end` | **Критично** |
| POST | **отсутствует** | `/api/sessions/answer` | **Критично** |

> Управление сессиями в архитектуре проекта должен взять на себя один из бэкендеров.
> Скорее всего это задача realtime-сервиса (он уже умеет `join`, `answer`, `start_session` через WS),
> либо quiz-core должен получить REST-эндпоинты для создания сессии и передачи PIN.

### Analytics / Reports

| Метод | URL (бэк) | URL (фронт) | Статус |
|---|---|---|---|
| GET | `/api/reports` | `/api/reports` | ✓ — но бэк возвращает `GameSession[]`, фронт ожидает `GameReport[]` (разные структуры) |
| GET | **отсутствует** | `/api/reports/${id}` | Эндпоинта нет |
| GET | `/api/reports/{id}/export` | `/api/reports/${id}/export` | ✓ URL совпадает |

---

## Критичные расхождения

### 1. Ответ `/auth/login` (критично)

Бэк возвращает:
```json
{ "token": "<JWT>" }
```

Фронт ожидает:
```json
{ "user": { "id": "...", "email": "...", "name": "...", "role": "teacher" }, "tokens": { "accessToken": "<JWT>" } }
```

После логина фронт пытается записать `response.tokens.accessToken` в localStorage — получит `undefined`.
Нужно либо адаптировать ответ бэка, либо добавить маппинг в `realApiService.ts`.

### 2. `quiz.id` — `int64` vs `string` (критично)

Фронт везде работает с `id` как строкой. Бэк отдаёт число. При передаче числа в URL (`/api/quizzes/42`) всё работает, но при сравнении `quiz.id === selectedId` (string vs number) логика сломается.

Нужно либо привести тип на фронте к `number`, либо попросить бэк вернуть строку.

### 3. Поля `timeLimitSeconds` / `timer_sec` (критично)

Фронт отправляет вопросы с полем `timeLimitSeconds`, бэк ожидает `timer_sec`. При сохранении квиза таймер не запишется.

### 4. Поля `isCorrect` / `is_correct` (критично)

Фронт отправляет `isCorrect`, бэк хранит `is_correct`. Правильные ответы не запишутся.

### 5. REST-сессии полностью отсутствуют (критично)

Весь flow студента (join → waiting → question → finished) и учителя (createSession → start → next) завязан на REST-эндпоинты сессий, которых в quiz-core нет. Это самый большой блок работы для коллеги.

### 6. `status` и `mode` у квизов (важно)

Фронт использует `status` для публикации квиза и `mode` для выбора режима. Бэк эти поля не хранит и не возвращает. После получения квиза с бэка `quiz.status` будет `undefined` — дашборд сломается.

---

## Что нужно согласовать с коллегами (backend-core)

1. **Привести ответ `/auth/login`** к структуре `{ user, tokens: { accessToken } }` или договориться об адаптере на фронте.
2. **Добавить поля в модель Quiz**: `status`, `mode`, `updated_at`.
3. **Переименовать поля** в snake_case → camelCase (или настроить JSON-теги в Go: `json:"timeLimitSeconds"`).
4. **Добавить `color` в Option** — нужен для отображения цветных вариантов ответов.
5. **Реализовать REST-эндпоинты сессий** — либо в quiz-core, либо отдельно.
6. **Уточнить**: кто создаёт комнату в realtime при `POST /api/sessions`? quiz-core должен вызвать `POST /api/rooms` на realtime-сервере с тем же PIN (см. `ws_integration.md`).

---

## Что уже работает без изменений

- Аутентификация (регистрация, смена пароля, сброс) — URL совпадают
- CRUD квизов — URL совпадают, структура частично совместима
- Экспорт CSV — реализован на бэке и подключён на фронте
- JWT-авторизация через `Authorization: Bearer` — схема одинаковая
