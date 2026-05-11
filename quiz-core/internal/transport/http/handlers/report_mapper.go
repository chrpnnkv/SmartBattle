package handlers

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/chrpnnkv/SmartBattle/internal/models"
)

// mapToDTO преобразует models.GameSession (со snapshot из realtime) в DTO,
// который ждёт фронтенд. Чистая функция (если не считать Quiz title — берётся
// из QuizService через метод-приёмник). Все вычисления — сортировка leaderboard,
// avgScore, проброс question_reports — здесь, чтобы handler-слой остался тонким.
func (h *ReportHandler) mapToDTO(s models.GameSession) GameReportDTO {
	dto := GameReportDTO{
		ID:              s.ID.String(),
		SessionID:       s.ID.String(),
		QuizID:          s.QuizID.String(),
		PlayedAt:        s.StartedAt,
		QuestionReports: []reportQuestionReport{},
		Leaderboard:     []reportLeaderboardEntry{},
	}
	if s.FinishedAt != nil {
		dto.PlayedAt = *s.FinishedAt
	}

	// Подтягиваем заголовок квиза из БД (если QuizService доступен).
	if h.quizService != nil {
		if quiz, err := h.quizService.GetQuizByID(s.QuizID); err == nil && quiz != nil {
			dto.QuizTitle = quiz.Title
		}
	}

	if len(s.ReportSnapshot) == 0 {
		return dto
	}

	var snap realtimeResultsSnapshot
	if err := json.Unmarshal(s.ReportSnapshot, &snap); err != nil {
		return dto
	}

	dto.ParticipantCount = len(snap.Results)

	// Сортируем по убыванию очков (как на финальном экране у студента).
	sortedResults := make([]struct {
		Name           string `json:"name"`
		Score          int    `json:"score"`
		CorrectAnswers int    `json:"correct_answers"`
		TotalQuestions int    `json:"total_questions"`
	}, len(snap.Results))
	copy(sortedResults, snap.Results)
	sort.SliceStable(sortedResults, func(i, j int) bool {
		return sortedResults[i].Score > sortedResults[j].Score
	})

	totalScore := 0
	for i, r := range sortedResults {
		totalScore += r.Score
		dto.Leaderboard = append(dto.Leaderboard, reportLeaderboardEntry{
			Rank:           i + 1,
			ID:             fmt.Sprintf("p_%d", i),
			Nickname:       r.Name,
			AvatarInitials: avatarInitials(r.Name),
			AvatarColor:    avatarColor(r.Name),
			Score:          r.Score,
			CorrectAnswers: r.CorrectAnswers,
			TotalQuestions: r.TotalQuestions,
			AnsweredCount:  r.CorrectAnswers,
		})
	}
	if dto.ParticipantCount > 0 {
		dto.AvgScore = float64(totalScore) / float64(dto.ParticipantCount)
	}
	if len(snap.QuestionReports) > 0 {
		dto.QuestionReports = snap.QuestionReports
	}

	return dto
}
