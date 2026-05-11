package handlers

import "time"

// reportLeaderboardEntry — позиция участника в финальной таблице.
type reportLeaderboardEntry struct {
	Rank           int    `json:"rank"`
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarInitials string `json:"avatarInitials"`
	AvatarColor    string `json:"avatarColor"`
	Score          int    `json:"score"`
	CorrectAnswers int    `json:"correctAnswers"`
	TotalQuestions int    `json:"totalQuestions"`
	// AnsweredCount оставлен для обратной совместимости со старым TS-типом
	// SessionParticipant; равен CorrectAnswers.
	AnsweredCount int `json:"answeredCount"`
}

// reportAnswerDistribution / reportFastestParticipant / reportQuestionReport —
// типизированные представления per-question-аналитики из realtime.
// JSON-теги в camelCase, чтобы напрямую отражать TS-тип фронтенда.
type reportAnswerDistribution struct {
	OptionID   string `json:"optionId"`
	OptionText string `json:"optionText"`
	Count      int    `json:"count"`
	IsCorrect  bool   `json:"isCorrect"`
	Color      string `json:"color"`
}

type reportFastestParticipant struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type reportQuestionReport struct {
	QuestionID                 string                     `json:"questionId"`
	QuestionText               string                     `json:"questionText"`
	CorrectPercent             int                        `json:"correctPercent"`
	AvgResponseTimeMs          int                        `json:"avgResponseTimeMs"`
	MostCommonWrongOptionID    string                     `json:"mostCommonWrongOptionId,omitempty"`
	MostCommonWrongOptionText  string                     `json:"mostCommonWrongOptionText,omitempty"`
	Distribution               []reportAnswerDistribution `json:"distribution"`
	FastestCorrectParticipants []reportFastestParticipant `json:"fastestCorrectParticipants"`
}

// GameReportDTO — структура, которую ждёт фронтенд (TS GameReport).
type GameReportDTO struct {
	ID               string                   `json:"id"`
	SessionID        string                   `json:"sessionId"`
	QuizID           string                   `json:"quizId"`
	QuizTitle        string                   `json:"quizTitle"`
	PlayedAt         time.Time                `json:"playedAt"`
	ParticipantCount int                      `json:"participantCount"`
	AvgScore         float64                  `json:"avgScore"`
	QuestionReports  []reportQuestionReport   `json:"questionReports"`
	Leaderboard      []reportLeaderboardEntry `json:"leaderboard"`
}

// realtimeResultsSnapshot — структура, которую кладёт realtime в reportSnapshot.
type realtimeResultsSnapshot struct {
	QuizID     string `json:"quiz_id"`
	RoomCode   string `json:"room_code"`
	Duration   int    `json:"duration_sec"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Results    []struct {
		Name           string `json:"name"`
		Score          int    `json:"score"`
		CorrectAnswers int    `json:"correct_answers"`
		TotalQuestions int    `json:"total_questions"`
	} `json:"results"`
	QuestionReports []reportQuestionReport `json:"question_reports"`
}
