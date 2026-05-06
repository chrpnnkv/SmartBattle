package handlers

import "time"

// SessionDTO — обогащённое представление сессии для фронтенда.
// Поля Mode/CurrentQuestionIndex/TotalQuestions/Participants собираются
// из БД + realtime-комнаты, чтобы фронтенд получал то, что описано в TS-типе GameSession.
type SessionDTO struct {
	ID                   string           `json:"id"`
	QuizID               string           `json:"quizId"`
	HostID               string           `json:"hostId"`
	PIN                  string           `json:"pin"`
	Status               string           `json:"status"`
	Mode                 string           `json:"mode"`
	StartedAt            time.Time        `json:"startedAt"`
	FinishedAt           *time.Time       `json:"finishedAt"`
	CurrentQuestionIndex int              `json:"currentQuestionIndex"`
	TotalQuestions       int              `json:"totalQuestions"`
	Participants         []SessionDTOPart `json:"participants"`
}

// SessionDTOPart — один участник в SessionDTO.Participants.
type SessionDTOPart struct {
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarInitials string `json:"avatarInitials"`
	AvatarColor    string `json:"avatarColor"`
	Score          int    `json:"score"`
	AnsweredCount  int    `json:"answeredCount"`
}
