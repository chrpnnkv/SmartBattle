package handlers

import (
	"testing"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
)

// mapQuizToDTO — на FE TS-тип Quiz объявлен с questionCount и estimatedMinutes.
// Раньше Core их не возвращал и Dashboard падал в NaN. Эти тесты фиксируют
// контракт расчёта обоих полей.

func TestMapQuizToDTO_EmptyQuiz(t *testing.T) {
	q := &models.Quiz{
		ID:        uuid.New(),
		TeacherID: uuid.New(),
		Title:     "Empty",
		Status:    "draft",
		Mode:      "teacher_paced",
	}

	dto := mapQuizToDTO(q)

	if dto.QuestionCount != 0 {
		t.Errorf("expected questionCount=0 for empty quiz, got %d", dto.QuestionCount)
	}
	if dto.EstimatedMinutes != 0 {
		t.Errorf("expected estimatedMinutes=0 for empty quiz, got %d", dto.EstimatedMinutes)
	}
	if dto.Settings == nil {
		t.Error("settings must be non-nil object even when quiz has no settings (FE expects {})")
	}
	if len(dto.Questions) != 0 {
		t.Errorf("expected 0 questions, got %d", len(dto.Questions))
	}
}

func TestMapQuizToDTO_QuestionCount(t *testing.T) {
	q := &models.Quiz{
		ID:        uuid.New(),
		TeacherID: uuid.New(),
		Title:     "Three",
		Questions: []models.Question{
			{ID: uuid.New(), TimerSec: 30},
			{ID: uuid.New(), TimerSec: 30},
			{ID: uuid.New(), TimerSec: 30},
		},
	}

	dto := mapQuizToDTO(q)

	if dto.QuestionCount != 3 {
		t.Errorf("expected questionCount=3, got %d", dto.QuestionCount)
	}
}

// EstimatedMinutes = ceil((Σ timerSec + 15с буфера * N) / 60).
// Для трёх вопросов по 30с: 3*30 + 3*15 = 135с = 3 минуты (округление вверх).
func TestMapQuizToDTO_EstimatedMinutes(t *testing.T) {
	q := &models.Quiz{
		ID:        uuid.New(),
		TeacherID: uuid.New(),
		Questions: []models.Question{
			{ID: uuid.New(), TimerSec: 30},
			{ID: uuid.New(), TimerSec: 30},
			{ID: uuid.New(), TimerSec: 30},
		},
	}

	dto := mapQuizToDTO(q)

	// 90 + 45 = 135 секунд → 3 минуты (ceil).
	if dto.EstimatedMinutes != 3 {
		t.Errorf("expected estimatedMinutes=3 for 3x30s questions, got %d", dto.EstimatedMinutes)
	}
}

// Округление вверх: 1 вопрос на 5 секунд + 15 буфер = 20 сек. Должно стать 1 минутой.
func TestMapQuizToDTO_EstimatedMinutesRoundsUp(t *testing.T) {
	q := &models.Quiz{
		ID: uuid.New(),
		Questions: []models.Question{
			{ID: uuid.New(), TimerSec: 5},
		},
	}
	dto := mapQuizToDTO(q)
	if dto.EstimatedMinutes != 1 {
		t.Errorf("expected estimatedMinutes=1 (round up), got %d", dto.EstimatedMinutes)
	}
}

func TestMapQuizToDTO_NilQuizReturnsEmpty(t *testing.T) {
	// nil не должен паниковать — обработчики могут вызвать с nil после ошибки выборки.
	dto := mapQuizToDTO(nil)
	if dto.ID != "" || dto.QuestionCount != 0 {
		t.Errorf("expected empty DTO for nil quiz, got %+v", dto)
	}
}

// Проверяем, что вопросы и опции пробрасываются с правильными ID и флагами.
// Без этого FE рисует пустые кружочки и не знает, какой ответ правильный.
func TestMapQuizToDTO_QuestionsAndOptions(t *testing.T) {
	qID := uuid.New()
	o1 := uuid.New()
	o2 := uuid.New()
	q := &models.Quiz{
		ID:        uuid.New(),
		TeacherID: uuid.New(),
		Questions: []models.Question{
			{
				ID:       qID,
				Type:     "multiple_choice",
				Text:     "What?",
				TimerSec: 30,
				Score:    100,
				Order:    0,
				Options: []models.Option{
					{ID: o1, Text: "Right", IsCorrect: true, Color: "green"},
					{ID: o2, Text: "Wrong", IsCorrect: false, Color: "red"},
				},
			},
		},
	}

	dto := mapQuizToDTO(q)

	if len(dto.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(dto.Questions))
	}
	got := dto.Questions[0]
	if got.ID != qID.String() {
		t.Errorf("question ID mismatch: got %s, want %s", got.ID, qID.String())
	}
	if got.TimeLimitSeconds != 30 {
		t.Errorf("expected timeLimitSeconds=30, got %d", got.TimeLimitSeconds)
	}
	if len(got.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(got.Options))
	}
	if !got.Options[0].IsCorrect || got.Options[1].IsCorrect {
		t.Errorf("isCorrect mapping is wrong: %+v", got.Options)
	}
	if got.Options[0].Color != "green" || got.Options[1].Color != "red" {
		t.Errorf("color mapping is wrong: %+v", got.Options)
	}
}
