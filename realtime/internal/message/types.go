package message

import "time"

const (
	// TypeJoin — вход в игровую комнату.
	TypeJoin = "join"
	// TypeAnswer — отправка ответа студентом.
	TypeAnswer = "answer"
	// TypeStartSession — команда преподавателя: начать сессию.
	TypeStartSession = "start_session"
	// TypeNextQuestion — команда преподавателя: перейти к следующему вопросу.
	TypeNextQuestion = "next_question"
	// TypeFinishSession — команда преподавателя: завершить сессию.
	TypeFinishSession = "finish_session"
	// TypePing — heartbeat от клиента.
	TypePing = "ping"
)

const (
	// TypeJoined — подтверждение входа в комнату.
	TypeJoined = "joined"
	// TypeParticipantJoined — уведомление о подключении нового участника.
	TypeParticipantJoined = "participant_joined"
	// TypeParticipantLeft — уведомление об отключении участника.
	TypeParticipantLeft = "participant_left"
	// TypeSessionStarted — сессия квиза запущена.
	TypeSessionStarted = "session_started"
	// TypeQuestion — текущий вопрос для всех участников.
	TypeQuestion = "question"
	// TypeAnswerReceived — уведомление преподавателю о поступившем ответе.
	TypeAnswerReceived = "answer_received"
	// TypeAnswerResult — результат ответа студенту (правильно/неправильно).
	TypeAnswerResult = "answer_result"
	// TypeQuestionResults — итоги по вопросу (для всех).
	TypeQuestionResults = "question_results"
	// TypeLeaderboard — текущий рейтинг участников.
	TypeLeaderboard = "leaderboard"
	// TypeSessionFinished — игровая сессия завершена.
	TypeSessionFinished = "session_finished"
	// TypeError — сообщение об ошибке.
	TypeError = "error"
	// TypePong — ответ на heartbeat.
	TypePong = "pong"
	// TypeWaiting — комната в режиме ожидания участников.
	TypeWaiting = "waiting"
)

// Роли участников
const (
	RoleTeacher = "teacher"
	RoleStudent = "student"
)

// Коды ошибок
const (
	ErrCodeInvalidToken      = "invalid_token"
	ErrCodeRoomNotFound      = "room_not_found"
	ErrCodeRoomFull          = "room_full"
	ErrCodeSessionNotActive  = "session_not_active"
	ErrCodeSessionAlreadyRun = "session_already_running"
	ErrCodeUnauthorized      = "unauthorized"
	ErrCodeInvalidMessage    = "invalid_message"
	ErrCodeRateLimitExceeded = "rate_limit_exceeded"
	ErrCodeAlreadyAnswered   = "already_answered"
	ErrCodeInternalError     = "internal_error"
)

// IncomingMessage — универсальный контейнер входящего WS-сообщения.
type IncomingMessage struct {
	Type        string `json:"type"`
	RoomCode    string `json:"room_code,omitempty"`
	Name        string `json:"name,omitempty"`
	Token       string `json:"token,omitempty"`
	QuestionID  string `json:"question_id,omitempty"`
	AnswerIndex int    `json:"answer_index"`
}

// OutgoingMessage — базовая обёртка исходящего сообщения.
type OutgoingMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload,omitempty"`
}

// JoinedPayload — ответ сервера на успешный join.
type JoinedPayload struct {
	RoomCode       string `json:"room_code"`
	Role           string `json:"role"`
	Name           string `json:"name"`
	QuizTitle      string `json:"quiz_title"`
	TotalQuestions int    `json:"total_questions"`
}

// ParticipantInfo — краткая информация об участнике.
type ParticipantInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// ParticipantJoinedPayload — уведомление о подключении участника.
type ParticipantJoinedPayload struct {
	Name         string            `json:"name"`
	Participants []ParticipantInfo `json:"participants"`
	TotalCount   int               `json:"total_count"`
}

// ParticipantLeftPayload — уведомление об отключении участника.
type ParticipantLeftPayload struct {
	Name       string `json:"name"`
	TotalCount int    `json:"total_count"`
}

// SessionStartedPayload — данные о запуске сессии.
type SessionStartedPayload struct {
	QuizTitle      string `json:"quiz_title"`
	TotalQuestions int    `json:"total_questions"`
}

// QuestionOption — вариант ответа на вопрос.
type QuestionOption struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// QuestionPayload — вопрос, рассылаемый всем участникам.
type QuestionPayload struct {
	QuestionID   string           `json:"question_id"`
	Index        int              `json:"index"`
	Total        int              `json:"total"`
	Text         string           `json:"text"`
	Options      []QuestionOption `json:"options"`
	TimeLimitSec int              `json:"time_limit_sec"`
	StartedAt    time.Time        `json:"started_at"`
}

// AnswerReceivedPayload — уведомление преподавателю о поступившем ответе.
type AnswerReceivedPayload struct {
	ParticipantName   string `json:"participant_name"`
	AnswersCount      int    `json:"answers_count"`
	TotalParticipants int    `json:"total_participants"`
}

// AnswerResultPayload — результат ответа студента.
type AnswerResultPayload struct {
	Correct      bool `json:"correct"`
	CorrectIndex int  `json:"correct_index"`
	Score        int  `json:"score"`
	TotalScore   int  `json:"total_score"`
}

// AnswerStat — статистика ответов по одному варианту.
type AnswerStat struct {
	OptionIndex int `json:"option_index"`
	Count       int `json:"count"`
}

// QuestionResultsPayload — итоги по вопросу.
type QuestionResultsPayload struct {
	QuestionID   string       `json:"question_id"`
	CorrectIndex int          `json:"correct_index"`
	Stats        []AnswerStat `json:"stats"`
	Leaderboard  []ScoreEntry `json:"leaderboard"`
}

// ScoreEntry — строка рейтинга.
type ScoreEntry struct {
	Rank  int    `json:"rank"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// LeaderboardPayload — текущий рейтинг.
type LeaderboardPayload struct {
	Entries []ScoreEntry `json:"entries"`
}

// ParticipantResult — итоговый результат участника.
type ParticipantResult struct {
	Name           string `json:"name"`
	Score          int    `json:"score"`
	CorrectAnswers int    `json:"correct_answers"`
	TotalQuestions int    `json:"total_questions"`
}

// SessionFinishedPayload — данные завершённой сессии.
type SessionFinishedPayload struct {
	QuizTitle string              `json:"quiz_title"`
	Results   []ParticipantResult `json:"results"`
	Duration  int                 `json:"duration_sec"`
}

// ErrorPayload — описание ошибки.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewError создаёт исходящее сообщение об ошибке.
func NewError(code, msg string) OutgoingMessage {
	return OutgoingMessage{
		Type:      TypeError,
		Timestamp: time.Now(),
		Payload:   ErrorPayload{Code: code, Message: msg},
	}
}

// New создаёт исходящее сообщение с заданным типом и payload.
func New(msgType string, payload interface{}) OutgoingMessage {
	return OutgoingMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// Pong возвращает pong-сообщение.
func Pong() OutgoingMessage {
	return OutgoingMessage{Type: TypePong, Timestamp: time.Now()}
}
