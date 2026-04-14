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
