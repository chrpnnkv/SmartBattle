package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// mapToDTO собирает GameReport из snapshot, который кладёт realtime.
// Эти тесты фиксируют контракт, который ждёт фронтенд:
//   - leaderboard отсортирован по score desc, ранги 1..N
//   - avgScore — среднее по всем участникам
//   - participantCount = len(results)
//   - Пустой snapshot → разумные дефолты, без паник
//   - question_reports пробрасываются как есть

func newTestReportHandler() *ReportHandler {
	// quizService=nil — мы не проверяем подгрузку quiz title в этих тестах.
	// mapToDTO защитно проверяет h.quizService != nil.
	return &ReportHandler{quizService: nil}
}

func makeSnapshot(t *testing.T, results []map[string]interface{}, questionReports []map[string]interface{}) datatypes.JSON {
	t.Helper()
	snap := map[string]interface{}{
		"quiz_id":      uuid.New().String(),
		"room_code":    "123456",
		"duration_sec": 100,
		"started_at":   "2026-05-02T10:00:00Z",
		"finished_at":  "2026-05-02T10:01:40Z",
		"results":      results,
	}
	if questionReports != nil {
		snap["question_reports"] = questionReports
	}
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	return datatypes.JSON(b)
}

func TestMapToDTO_EmptySnapshot(t *testing.T) {
	h := newTestReportHandler()
	s := models.GameSession{
		ID:        uuid.New(),
		QuizID:    uuid.New(),
		HostID:    uuid.New(),
		PIN:       "ABC123",
		Status:    "finished",
		StartedAt: time.Now(),
	}
	dto := h.mapToDTO(s)

	if dto.ParticipantCount != 0 {
		t.Errorf("expected participantCount=0, got %d", dto.ParticipantCount)
	}
	if dto.AvgScore != 0 {
		t.Errorf("expected avgScore=0, got %f", dto.AvgScore)
	}
	// Должен быть пустой массив, а НЕ nil — иначе фронт видит null и крэшится.
	if dto.Leaderboard == nil {
		t.Error("expected leaderboard=[], got nil")
	}
	if dto.QuestionReports == nil {
		t.Error("expected questionReports=[], got nil")
	}
}

func TestMapToDTO_LeaderboardSortedByScoreDesc(t *testing.T) {
	h := newTestReportHandler()
	results := []map[string]interface{}{
		{"name": "Charlie", "score": 100, "correct_answers": 1, "total_questions": 3},
		{"name": "Alice", "score": 300, "correct_answers": 3, "total_questions": 3},
		{"name": "Bob", "score": 200, "correct_answers": 2, "total_questions": 3},
	}

	s := models.GameSession{
		ID:             uuid.New(),
		QuizID:         uuid.New(),
		ReportSnapshot: makeSnapshot(t, results, nil),
	}
	dto := h.mapToDTO(s)

	if len(dto.Leaderboard) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(dto.Leaderboard))
	}
	if dto.Leaderboard[0].Nickname != "Alice" || dto.Leaderboard[0].Score != 300 || dto.Leaderboard[0].Rank != 1 {
		t.Errorf("rank 1 wrong: %+v", dto.Leaderboard[0])
	}
	if dto.Leaderboard[1].Nickname != "Bob" || dto.Leaderboard[1].Rank != 2 {
		t.Errorf("rank 2 wrong: %+v", dto.Leaderboard[1])
	}
	if dto.Leaderboard[2].Nickname != "Charlie" || dto.Leaderboard[2].Rank != 3 {
		t.Errorf("rank 3 wrong: %+v", dto.Leaderboard[2])
	}
}

func TestMapToDTO_AvgScore(t *testing.T) {
	h := newTestReportHandler()
	results := []map[string]interface{}{
		{"name": "A", "score": 100, "correct_answers": 1, "total_questions": 2},
		{"name": "B", "score": 200, "correct_answers": 2, "total_questions": 2},
		{"name": "C", "score": 300, "correct_answers": 2, "total_questions": 2},
	}
	s := models.GameSession{
		ID:             uuid.New(),
		QuizID:         uuid.New(),
		ReportSnapshot: makeSnapshot(t, results, nil),
	}
	dto := h.mapToDTO(s)

	if dto.ParticipantCount != 3 {
		t.Errorf("expected participantCount=3, got %d", dto.ParticipantCount)
	}
	if dto.AvgScore != 200.0 {
		t.Errorf("expected avgScore=200, got %f", dto.AvgScore)
	}
}

// CorrectAnswers/TotalQuestions из snapshot должны попасть в leaderboard,
// чтобы FE мог посчитать процент точности.
func TestMapToDTO_LeaderboardCorrectAndTotal(t *testing.T) {
	h := newTestReportHandler()
	results := []map[string]interface{}{
		{"name": "Alice", "score": 250, "correct_answers": 5, "total_questions": 10},
	}
	s := models.GameSession{
		ID:             uuid.New(),
		QuizID:         uuid.New(),
		ReportSnapshot: makeSnapshot(t, results, nil),
	}
	dto := h.mapToDTO(s)

	if len(dto.Leaderboard) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(dto.Leaderboard))
	}
	got := dto.Leaderboard[0]
	if got.CorrectAnswers != 5 || got.TotalQuestions != 10 {
		t.Errorf("expected 5/10, got %d/%d", got.CorrectAnswers, got.TotalQuestions)
	}
	// answeredCount оставлен для обратной совместимости и должен равняться correctAnswers.
	if got.AnsweredCount != 5 {
		t.Errorf("expected answeredCount=5 (= correctAnswers), got %d", got.AnsweredCount)
	}
}

// Если snapshot содержит question_reports — Core должен пробросить их в DTO.
func TestMapToDTO_QuestionReportsFlowThrough(t *testing.T) {
	h := newTestReportHandler()
	qReports := []map[string]interface{}{
		{
			"questionId":        "q1",
			"questionText":      "Sample?",
			"correctPercent":    75,
			"avgResponseTimeMs": 1200,
			"distribution": []map[string]interface{}{
				{"optionId": "o1", "optionText": "A", "count": 3, "isCorrect": true, "color": "red"},
			},
			"fastestCorrectParticipants": []map[string]interface{}{
				{"id": "p1", "nickname": "Alice"},
			},
		},
	}
	results := []map[string]interface{}{
		{"name": "Alice", "score": 100, "correct_answers": 1, "total_questions": 1},
	}

	s := models.GameSession{
		ID:             uuid.New(),
		QuizID:         uuid.New(),
		ReportSnapshot: makeSnapshot(t, results, qReports),
	}
	dto := h.mapToDTO(s)

	if len(dto.QuestionReports) != 1 {
		t.Fatalf("expected 1 question report, got %d", len(dto.QuestionReports))
	}
	got := dto.QuestionReports[0]
	if got.QuestionID != "q1" {
		t.Errorf("expected questionId=q1, got %s", got.QuestionID)
	}
	if got.CorrectPercent != 75 {
		t.Errorf("expected correctPercent=75, got %d", got.CorrectPercent)
	}
	if len(got.Distribution) != 1 || got.Distribution[0].Color != "red" {
		t.Errorf("distribution not parsed: %+v", got.Distribution)
	}
	if len(got.FastestCorrectParticipants) != 1 || got.FastestCorrectParticipants[0].Nickname != "Alice" {
		t.Errorf("fastestCorrectParticipants not parsed: %+v", got.FastestCorrectParticipants)
	}
}

// PlayedAt: если в БД finishedAt установлен, он используется как playedAt;
// иначе fallback на startedAt.
func TestMapToDTO_PlayedAtFallback(t *testing.T) {
	h := newTestReportHandler()
	started := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	// Без finishedAt — playedAt = startedAt.
	s := models.GameSession{
		ID:        uuid.New(),
		QuizID:    uuid.New(),
		StartedAt: started,
	}
	dto := h.mapToDTO(s)
	if !dto.PlayedAt.Equal(started) {
		t.Errorf("expected playedAt=%v, got %v", started, dto.PlayedAt)
	}

	// С finishedAt — берём finishedAt.
	finished := started.Add(2 * time.Minute)
	s2 := models.GameSession{
		ID:         uuid.New(),
		QuizID:     uuid.New(),
		StartedAt:  started,
		FinishedAt: &finished,
	}
	dto2 := h.mapToDTO(s2)
	if !dto2.PlayedAt.Equal(finished) {
		t.Errorf("expected playedAt=%v, got %v", finished, dto2.PlayedAt)
	}
}
