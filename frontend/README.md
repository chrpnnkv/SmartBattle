# Smart Battle - Frontend

Веб-приложение для проведения академических квизов в реальном времени. Преподаватель создаёт квиз и управляет сессией, студенты подключаются по PIN-коду и отвечают на вопросы

## Стек

- **React 18** + **TypeScript**
- **Redux Toolkit** - управление состоянием
- **React Router v6** - маршрутизация
- **CSS Modules** - стилизация
- **Vite** - сборщик
- **WebSocket** - real-time взаимодействие

## Запуск локально

### Требования

- Node.js 18+
- npm 9+

### Установка и запуск

```bash
# Установить зависимости
npm install

# Создать файл окружения
cp .env.example .env

# Запустить в режиме разработки
npm run dev
```

Приложение будет доступно на `http://localhost:5173`.

По умолчанию используется mock API

### Сборка

```bash
npm run build
```

## Переменные окружения

Создайте файл `.env` на основе `.env.example`:

| Переменная | Описание | По умолчанию |
|---|---|---|
| `VITE_USE_MOCK` | Использовать mock API | `true` |
| `VITE_API_URL` | URL бэкенда | `http://localhost:8080` |
| `VITE_WS_URL` | URL WebSocket | `ws://localhost:8080` |

## Структура проекта

```
src/
├── api/
│   ├── IApiService.ts        # Абстрактный интерфейс API
│   ├── mock/                 # Mock реализация (localStorage)
│   ├── real/                 # Реальный HTTP клиент
│   ├── wsService.ts          # WebSocket сервис (mock + real)
│   └── index.ts              # Переключатель mock/real
├── components/
│   ├── layout/               # AppLayout, AuthLayout
│   └── ui/                   # Button, Input, Modal, Badge...
├── pages/                    # Страницы приложения
├── store/
│   └── slices/               # authSlice, quizSlice, sessionSlice
├── types/                    # TypeScript типы
└── router/                   # AppRouter, ProtectedRoute
```

## Подключение бэкенда

1. Установить в `.env`:
   ```
   VITE_USE_MOCK=false
   VITE_API_URL=http://адрес-бэка:порт
   VITE_WS_URL=ws://адрес-бэка:порт
   ```

## Роли и страницы

### Преподаватель
| Страница | Путь |
|---|---|
| Вход / Регистрация | `/login`, `/register` |
| Личный кабинет | `/dashboard` |
| Конструктор квизов | `/quiz/new`, `/quiz/:id/edit` |
| Проведение квиза | `/session/:id/analytics` |
| Отчёты | `/reports` |

### Студент
| Страница | Путь |
|---|---|
| Главная | `/` |
| Вход по PIN | `/join` |
| Комната ожидания | `/session/:id/waiting` |
| Вопрос | `/session/:id/question` |
| Результаты | `/session/:id/finished` |

## Режимы проведения квиза

- **Teacher-paced** - преподаватель вручную переключает вопросы
- **Student-paced** - переход к следующему вопросу автоматически, когда все ответили или истекло время
