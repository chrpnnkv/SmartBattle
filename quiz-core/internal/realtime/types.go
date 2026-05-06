// Package realtime содержит клиента для realtime-сервиса и типы wire-протокола
// между quiz-core (этот сервис) и backend-realtime.
//
// Назначение пакета — изолировать всю интеграцию с realtime в одном месте.
// Сервисный слой (internal/service) видит только интерфейс Client и доменные
// модели; HTTP-вызовы, JWT-подпись, JSON-форматы лежат здесь.
package realtime

import "time"

// Option — вариант ответа, отправляемый в realtime при создании комнаты.
type Option struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Color     string `json:"color"`
}

// Question — вопрос квиза, отправляемый в realtime.
type Question struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	ImageURL     string   `json:"image_url,omitempty"`
	Score        int      `json:"score"`
	Options      []Option `json:"options"`
	TimeLimitSec int      `json:"time_limit_sec"`
}

// CreateRoomRequest — тело POST /api/rooms на realtime-сервере.
type CreateRoomRequest struct {
	QuizID    string     `json:"quiz_id"`
	QuizTitle string     `json:"quiz_title"`
	QuizMode  string     `json:"quiz_mode,omitempty"`
	Questions []Question `json:"questions"`
}

// CreateRoomResponse — успешный ответ от POST /api/rooms.
type CreateRoomResponse struct {
	RoomCode string `json:"room_code"`
	Error    string `json:"error"`
}

// RoomInfo — ответ GET /api/rooms/:code (используется для self-healing
// восстановления сессии, потерянной в БД).
type RoomInfo struct {
	RoomCode     string `json:"room_code"`
	QuizID       string `json:"quiz_id"`
	QuizMode     string `json:"quiz_mode"`
	HostID       string `json:"host_id"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	Participants int    `json:"participants"`
}

// Participant — участник комнаты в формате, который отдаёт realtime.
type Participant struct {
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarInitials string `json:"avatarInitials"`
	AvatarColor    string `json:"avatarColor"`
	Score          int    `json:"score"`
	AnsweredCount  int    `json:"answeredCount"`
}

// RoomParticipants — ответ GET /api/rooms/:code/participants.
type RoomParticipants struct {
	Participants         []Participant `json:"participants"`
	CurrentQuestionIndex int           `json:"current_question_index"`
}

// Таймауты по умолчанию — заметные константы вместо магических чисел.
const (
	DefaultCreateRoomTimeout   = 5 * time.Second
	DefaultRoomInfoTimeout     = 5 * time.Second
	DefaultParticipantsTimeout = 5 * time.Second
)
