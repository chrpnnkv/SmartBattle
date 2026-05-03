package handlers

import (
	"encoding/json"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Per-question 15с буфера на показ результата + переход. Используется
// при расчёте estimatedMinutes.
const questionBufferSec = 15

// Стандартный формат timestamp для FE (ISO 8601 c таймзоной).
const isoLayout = "2006-01-02T15:04:05Z07:00"

// mapQuizInputToModel принимает QuizInput и собирает models.Quiz, готовый
// для сохранения через GORM. UUID для quiz/questions/options генерируется,
// если не пришёл с фронта.
func mapQuizInputToModel(input QuizInput, teacherID uuid.UUID) *models.Quiz {
	quiz := &models.Quiz{
		TeacherID:   teacherID,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		Mode:        input.Mode,
	}

	settingsBytes, _ := json.Marshal(input.Settings)
	quiz.Settings = datatypes.JSON(settingsBytes)

	for i, qIn := range input.Questions {
		qID, err := uuid.Parse(qIn.ID)
		if err != nil {
			qID = uuid.New()
		}

		correctAnswersBytes, _ := json.Marshal(qIn.CorrectTextAnswers)

		q := models.Question{
			ID:                 qID,
			Type:               qIn.Type,
			Text:               qIn.Text,
			ImageURL:           qIn.ImageURL,
			TimerSec:           qIn.TimeLimitSeconds,
			Score:              qIn.Score,
			Order:              i,
			CorrectTextAnswers: datatypes.JSON(correctAnswersBytes),
		}

		for _, oIn := range qIn.Options {
			oID, err := uuid.Parse(oIn.ID)
			if err != nil {
				oID = uuid.New()
			}
			q.Options = append(q.Options, models.Option{
				ID:        oID,
				Text:      oIn.Text,
				IsCorrect: oIn.IsCorrect,
				Color:     oIn.Color,
			})
		}
		quiz.Questions = append(quiz.Questions, q)
	}
	return quiz
}

// mapQuizToDTO преобразует models.Quiz в DTO для фронта. Чистая функция.
// Считает questionCount и estimatedMinutes (производные поля, ждущиеся TS-типом).
func mapQuizToDTO(q *models.Quiz) QuizDTO {
	if q == nil {
		return QuizDTO{}
	}

	dto := QuizDTO{
		ID:          q.ID.String(),
		Title:       q.Title,
		Description: q.Description,
		Status:      q.Status,
		Mode:        q.Mode,
		AuthorID:    q.TeacherID.String(),
		CreatedAt:   q.CreatedAt.UTC().Format(isoLayout),
		UpdatedAt:   q.UpdatedAt.UTC().Format(isoLayout),
		Questions:   make([]QuestionDTO, 0, len(q.Questions)),
	}

	if len(q.Settings) > 0 {
		var settings map[string]interface{}
		if err := json.Unmarshal(q.Settings, &settings); err == nil {
			dto.Settings = settings
		}
	}
	if dto.Settings == nil {
		dto.Settings = map[string]interface{}{}
	}

	totalSec := 0
	for _, qq := range q.Questions {
		opts := make([]OptionDTO, 0, len(qq.Options))
		for _, o := range qq.Options {
			opts = append(opts, OptionDTO{
				ID:         o.ID.String(),
				QuestionID: o.QuestionID.String(),
				Text:       o.Text,
				IsCorrect:  o.IsCorrect,
				Color:      o.Color,
			})
		}
		var correctAnswers []string
		if len(qq.CorrectTextAnswers) > 0 {
			_ = json.Unmarshal(qq.CorrectTextAnswers, &correctAnswers)
		}
		dto.Questions = append(dto.Questions, QuestionDTO{
			ID:                 qq.ID.String(),
			QuizID:             qq.QuizID.String(),
			Type:               qq.Type,
			Text:               qq.Text,
			ImageURL:           qq.ImageURL,
			TimeLimitSeconds:   qq.TimerSec,
			Score:              qq.Score,
			Order:              qq.Order,
			Options:            opts,
			CorrectTextAnswers: correctAnswers,
		})
		totalSec += qq.TimerSec
	}
	totalSec += len(q.Questions) * questionBufferSec

	dto.QuestionCount = len(q.Questions)
	if totalSec > 0 {
		dto.EstimatedMinutes = (totalSec + 59) / 60
	}
	return dto
}

func mapQuizzesToDTO(quizzes []models.Quiz) []QuizDTO {
	out := make([]QuizDTO, 0, len(quizzes))
	for i := range quizzes {
		out = append(out, mapQuizToDTO(&quizzes[i]))
	}
	return out
}
