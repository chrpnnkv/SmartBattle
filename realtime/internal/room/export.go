package room

import (
	"time"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/core"
)

// GetCode возвращает код комнаты.
func (r *Room) GetCode() string { return r.Code }

// GetQuizID возвращает ID квиза.
func (r *Room) GetQuizID() string { return r.QuizID }

// GetStartedAt возвращает время начала сессии.
func (r *Room) GetStartedAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.StartedAt
}

// GetFinishedAt возвращает время завершения сессии.
func (r *Room) GetFinishedAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FinishedAt
}

// GetCoreResults возвращает результаты в формате, нужном core-клиенту.
func (r *Room) GetCoreResults() []core.ParticipantResult {
	results := r.GetResults()
	out := make([]core.ParticipantResult, len(results))
	for i, res := range results {
		out[i] = core.ParticipantResult{
			Name:           res.Name,
			Score:          res.Score,
			CorrectAnswers: res.CorrectAnswers,
			TotalQuestions: res.TotalQuestions,
		}
	}
	return out
}

// GetCoreQuestionReports — собирает накопленные отчёты по вопросам в формат,
// который понимает backend-core (camelCase JSON-теги совпадают с TS-типом QuestionReport).
func (r *Room) GetCoreQuestionReports() []core.QuestionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]core.QuestionReport, 0, len(r.QuestionReports))
	for _, qr := range r.QuestionReports {
		dist := make([]core.AnswerDistribution, 0, len(qr.Distribution))
		for _, d := range qr.Distribution {
			dist = append(dist, core.AnswerDistribution{
				OptionID:   d.OptionID,
				OptionText: d.OptionText,
				Count:      d.Count,
				IsCorrect:  d.IsCorrect,
				Color:      d.Color,
			})
		}
		fastest := make([]core.ParticipantShort, 0, len(qr.FastestCorrectParticipants))
		for _, p := range qr.FastestCorrectParticipants {
			fastest = append(fastest, core.ParticipantShort{ID: p.ID, Nickname: p.Nickname})
		}
		out = append(out, core.QuestionReport{
			QuestionID:                 qr.QuestionID,
			QuestionText:               qr.QuestionText,
			CorrectPercent:             qr.CorrectPercent,
			AvgResponseTimeMs:          qr.AvgResponseTimeMs,
			MostCommonWrongOptionID:    qr.MostCommonWrongOptionID,
			MostCommonWrongOptionText:  qr.MostCommonWrongOptionText,
			Distribution:               dist,
			FastestCorrectParticipants: fastest,
		})
	}
	return out
}
