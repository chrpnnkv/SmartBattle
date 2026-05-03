package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ResultsPayload — тело запроса передачи итогов сессии.
type ResultsPayload struct {
	QuizID     string              `json:"quiz_id"`
	RoomCode   string              `json:"room_code"`
	StartedAt  time.Time           `json:"started_at"`
	FinishedAt time.Time           `json:"finished_at"`
	Duration   int                 `json:"duration_sec"`
	Results    []ParticipantResult `json:"results"`
	// QuestionReports намеренно в camelCase + ключ snake-case массива:
	// в backend-core поле читается как question_reports.
	QuestionReports []QuestionReport `json:"question_reports,omitempty"`
}

// ParticipantResult — результат одного участника.
type ParticipantResult struct {
	Name           string `json:"name"`
	Score          int    `json:"score"`
	CorrectAnswers int    `json:"correct_answers"`
	TotalQuestions int    `json:"total_questions"`
}

// QuestionReport — детальный отчёт по одному вопросу для последующего показа на /reports.
// JSON-теги в camelCase, чтобы напрямую совпадать с TS-типом QuestionReport фронтенда.
type QuestionReport struct {
	QuestionID                 string               `json:"questionId"`
	QuestionText               string               `json:"questionText"`
	CorrectPercent             int                  `json:"correctPercent"`
	AvgResponseTimeMs          int                  `json:"avgResponseTimeMs"`
	MostCommonWrongOptionID    string               `json:"mostCommonWrongOptionId,omitempty"`
	MostCommonWrongOptionText  string               `json:"mostCommonWrongOptionText,omitempty"`
	Distribution               []AnswerDistribution `json:"distribution"`
	FastestCorrectParticipants []ParticipantShort   `json:"fastestCorrectParticipants"`
}

// AnswerDistribution — статистика ответов по варианту.
type AnswerDistribution struct {
	OptionID   string `json:"optionId"`
	OptionText string `json:"optionText"`
	Count      int    `json:"count"`
	IsCorrect  bool   `json:"isCorrect"`
	Color      string `json:"color"`
}

// ParticipantShort — короткая запись участника для списка лидеров вопроса.
type ParticipantShort struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

// QuizData — данные квиза, получаемые из backend-core.
type QuizData struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

// Question — вопрос из backend-core.
type Question struct {
	ID           string   `json:"id"`
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	TimeLimitSec int      `json:"time_limit_sec"`
}

// ResultsSaver — интерфейс для передачи результатов.
type ResultsSaver interface {
	SaveResults(r SessionResult) error
}

// SessionResult — результаты сессии.
type SessionResult interface {
	GetCode() string
	GetQuizID() string
	GetStartedAt() time.Time
	GetFinishedAt() time.Time
	GetResults() interface{}
}

// Client — HTTP-клиент для backend-core.
type Client struct {
	baseURL        string
	httpClient     *http.Client
	internalSecret string
	logger         *slog.Logger
}

// New создаёт новый Client для backend-core.
func New(baseURL, internalSecret string, timeout time.Duration, logger *slog.Logger) *Client {
	return &Client{
		baseURL:        baseURL,
		internalSecret: internalSecret,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger.With("component", "core_client"),
	}
}

func (c *Client) SaveResultsPayload(payload ResultsPayload) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
	defer cancel()

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации результатов: %w", err)
	}

	url := c.baseURL + "/internal/reports"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.internalSecret != "" {
		req.Header.Set("X-Internal-Secret", c.internalSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка запроса к backend-core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("backend-core вернул статус %d", resp.StatusCode)
	}

	c.logger.Info("результаты успешно переданы в backend-core",
		"room_code", payload.RoomCode,
		"quiz_id", payload.QuizID,
		"participants", len(payload.Results),
		"question_reports", len(payload.QuestionReports),
	)
	return nil
}

// SaveResultsPayloadWithRetry — повторная отправка с экспоненциальной паузой,
// чтобы кратковременная недоступность backend-core не приводила к потере отчёта.
func (c *Client) SaveResultsPayloadWithRetry(payload ResultsPayload) error {
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.SaveResultsPayload(payload); err != nil {
			lastErr = err
			c.logger.Warn("не удалось отправить отчёт, повторяю",
				"attempt", attempt,
				"max", maxAttempts,
				"error", err,
				"room_code", payload.RoomCode,
			)
			if attempt < maxAttempts {
				// 1s, 2s, 4s
				time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
			}
			continue
		}
		return nil
	}
	c.logger.Error("отчёт окончательно не доставлен в backend-core",
		"room_code", payload.RoomCode,
		"quiz_id", payload.QuizID,
		"error", lastErr,
	)
	return lastErr
}

// GetQuiz получает данные квиза из backend-core по ID.
func (c *Client) GetQuiz(ctx context.Context, quizID string) (*QuizData, error) {
	url := fmt.Sprintf("%s/internal/quizzes/%s", c.baseURL, quizID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.internalSecret != "" {
		req.Header.Set("X-Internal-Secret", c.internalSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к backend-core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("квиз %s не найден", quizID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend-core вернул статус %d", resp.StatusCode)
	}

	var quiz QuizData
	if err := json.NewDecoder(resp.Body).Decode(&quiz); err != nil {
		return nil, fmt.Errorf("ошибка десериализации квиза: %w", err)
	}
	return &quiz, nil
}

// SaveResults — адаптер для вызова из room.SetOnFinish.
func (c *Client) SaveResults(r roomFinisher) error {
	payload := ResultsPayload{
		QuizID:          r.GetQuizID(),
		RoomCode:        r.GetCode(),
		StartedAt:       r.GetStartedAt(),
		FinishedAt:      r.GetFinishedAt(),
		Duration:        int(r.GetFinishedAt().Sub(r.GetStartedAt()).Seconds()),
		Results:         r.GetCoreResults(),
		QuestionReports: r.GetCoreQuestionReports(),
	}
	return c.SaveResultsPayloadWithRetry(payload)
}

// roomFinisher — минимальный интерфейс для извлечения данных из комнаты.
type roomFinisher interface {
	GetQuizID() string
	GetCode() string
	GetStartedAt() time.Time
	GetFinishedAt() time.Time
	GetCoreResults() []ParticipantResult
	GetCoreQuestionReports() []QuestionReport
}
