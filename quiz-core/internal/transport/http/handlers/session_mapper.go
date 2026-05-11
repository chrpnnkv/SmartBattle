package handlers

import (
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/realtime"
	"github.com/chrpnnkv/SmartBattle/internal/service"
)

// buildSessionDTO собирает SessionDTO для фронтенда.
// totalQuestions берётся из квиза; participants/currentQuestionIndex —
// best-effort из realtime через SessionService.FetchLiveRoom.
func buildSessionDTO(svc *service.SessionService, quizSvc *service.QuizService, s *models.GameSession) SessionDTO {
	dto := SessionDTO{
		ID:           s.ID.String(),
		QuizID:       s.QuizID.String(),
		HostID:       s.HostID.String(),
		PIN:          s.PIN,
		Status:       s.Status,
		Mode:         s.Mode,
		StartedAt:    s.StartedAt,
		FinishedAt:   s.FinishedAt,
		Participants: []SessionDTOPart{},
	}

	if quizSvc != nil {
		if quiz, err := quizSvc.GetQuizByID(s.QuizID); err == nil && quiz != nil {
			dto.TotalQuestions = len(quiz.Questions)
		}
	}

	if s.Status == "waiting" || s.Status == "active" {
		if live, ok := svc.FetchLiveRoom(s.PIN); ok {
			dto.Participants = mapRealtimeParticipants(live.Participants)
			dto.CurrentQuestionIndex = live.CurrentQuestionIndex
		}
	}
	return dto
}

func mapRealtimeParticipants(in []realtime.Participant) []SessionDTOPart {
	out := make([]SessionDTOPart, 0, len(in))
	for _, p := range in {
		out = append(out, SessionDTOPart{
			ID:             p.ID,
			Nickname:       p.Nickname,
			AvatarInitials: p.AvatarInitials,
			AvatarColor:    p.AvatarColor,
			Score:          p.Score,
			AnsweredCount:  p.AnsweredCount,
		})
	}
	return out
}
