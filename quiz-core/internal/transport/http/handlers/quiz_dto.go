package handlers

// Входные DTO (создание / обновление).

type OptionInput struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
	Color     string `json:"color"`
}

type QuestionInput struct {
	ID                 string        `json:"id"`
	Type               string        `json:"type"`
	Text               string        `json:"text"`
	ImageURL           string        `json:"imageUrl"`
	TimeLimitSeconds   int           `json:"timeLimitSeconds"`
	Score              int           `json:"score"`
	Options            []OptionInput `json:"options"`
	CorrectTextAnswers []string      `json:"correctTextAnswers"`
}

// QuizInput — общий ввод для UpdateQuiz (все поля необязательны).
type QuizInput struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Mode        string                 `json:"mode"`
	Settings    map[string]interface{} `json:"settings"`
	Questions   []QuestionInput        `json:"questions"`
}

// CreateQuizInput используется только для CreateQuiz и требует Title.
type CreateQuizInput struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Mode        string                 `json:"mode"`
	Settings    map[string]interface{} `json:"settings"`
	Questions   []QuestionInput        `json:"questions"`
}

// Выходные DTO — то, что фронт ждёт по контракту TS-типа Quiz.
// QuestionCount и EstimatedMinutes — производные поля, считаются мапером.

type QuizDTO struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Status           string                 `json:"status"`
	Mode             string                 `json:"mode"`
	Settings         map[string]interface{} `json:"settings"`
	AuthorID         string                 `json:"authorId"`
	CreatedAt        string                 `json:"createdAt"`
	UpdatedAt        string                 `json:"updatedAt"`
	Questions        []QuestionDTO          `json:"questions"`
	QuestionCount    int                    `json:"questionCount"`
	EstimatedMinutes int                    `json:"estimatedMinutes"`
}

type QuestionDTO struct {
	ID                 string      `json:"id"`
	QuizID             string      `json:"quizId"`
	Type               string      `json:"type"`
	Text               string      `json:"text"`
	ImageURL           string      `json:"imageUrl,omitempty"`
	TimeLimitSeconds   int         `json:"timeLimitSeconds"`
	Score              int         `json:"score"`
	Order              int         `json:"order"`
	Options            []OptionDTO `json:"options"`
	CorrectTextAnswers []string    `json:"correctTextAnswers,omitempty"`
}

type OptionDTO struct {
	ID         string `json:"id"`
	QuestionID string `json:"questionId"`
	Text       string `json:"text"`
	IsCorrect  bool   `json:"isCorrect"`
	Color      string `json:"color"`
}
