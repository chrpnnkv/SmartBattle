package realtime

import "time"

type Option struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Color     string `json:"color"`
}

type Question struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	ImageURL     string   `json:"image_url,omitempty"`
	Score        int      `json:"score"`
	Options      []Option `json:"options"`
	TimeLimitSec int      `json:"time_limit_sec"`
}

type CreateRoomRequest struct {
	QuizID    string     `json:"quiz_id"`
	QuizTitle string     `json:"quiz_title"`
	QuizMode  string     `json:"quiz_mode,omitempty"`
	Questions []Question `json:"questions"`
}

type CreateRoomResponse struct {
	RoomCode string `json:"room_code"`
	Error    string `json:"error"`
}

type RoomInfo struct {
	RoomCode     string `json:"room_code"`
	QuizID       string `json:"quiz_id"`
	QuizMode     string `json:"quiz_mode"`
	HostID       string `json:"host_id"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	Participants int    `json:"participants"`
}

type Participant struct {
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarInitials string `json:"avatarInitials"`
	AvatarColor    string `json:"avatarColor"`
	Score          int    `json:"score"`
	AnsweredCount  int    `json:"answeredCount"`
}

type RoomParticipants struct {
	Participants         []Participant `json:"participants"`
	CurrentQuestionIndex int           `json:"current_question_index"`
}

const (
	DefaultCreateRoomTimeout = 5 * time.Second
)
