package room

import (
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/client"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
)

// Status — статус игровой комнаты.
type Status string

const (
	StatusWaiting  Status = "waiting"
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

// Question — вопрос квиза.
type Question struct {
	ID           string
	Text         string
	Options      []string
	CorrectIndex int
	TimeLimitSec int
}

// Answer — ответ участника.
type Answer struct {
	AnswerIndex int
	AnsweredAt  time.Time
	IsCorrect   bool
	Score       int
}

// Participant — участник сессии.
type Participant struct {
	Client     *client.Client
	Name       string
	TotalScore int
	Answers    map[string]Answer
}

// Room — игровая комната.
type Room struct {
	mu   sync.RWMutex
	Code string

	Status    Status
	QuizID    string
	QuizTitle string
	Questions []Question

	TeacherID    string
	Participants map[string]*Participant

	CurrentQuestionIndex int
	QuestionStartedAt    time.Time

	StartedAt  time.Time
	FinishedAt time.Time

	cfg    RoomConfig
	logger *slog.Logger

	onFinish func(r *Room)
}

// RoomConfig — параметры комнаты.
type RoomConfig struct {
	MaxParticipants        int
	DefaultQuestionTimeSec int
}

// New создаёт новую игровую комнату.
func New(code, quizID, quizTitle string, questions []Question, cfg RoomConfig, logger *slog.Logger) *Room {
	return &Room{
		Code:         code,
		Status:       StatusWaiting,
		QuizID:       quizID,
		QuizTitle:    quizTitle,
		Questions:    questions,
		Participants: make(map[string]*Participant),
		cfg:          cfg,
		logger:       logger.With("component", "room", "room_code", code),
	}
}

// SetOnFinish устанавливает callback на завершение сессии.
func (r *Room) SetOnFinish(fn func(r *Room)) {
	r.mu.Lock()
	r.onFinish = fn
	r.mu.Unlock()
}

// AddClient добавляет клиента в комнату.
func (r *Room) AddClient(c *client.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StatusFinished {
		return errors.New("сессия уже завершена")
	}
	if c.Role == message.RoleStudent && len(r.students()) >= r.cfg.MaxParticipants {
		return errors.New("комната заполнена")
	}
	if c.Role == message.RoleTeacher {
		r.TeacherID = c.ID
	}

	r.Participants[c.ID] = &Participant{
		Client:  c,
		Name:    c.Name,
		Answers: make(map[string]Answer),
	}

	r.logger.Info("участник подключился", "client_id", c.ID, "name", c.Name, "role", c.Role)
	return nil
}

// RemoveClient удаляет клиента из комнаты.
func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	p, ok := r.Participants[clientID]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.Participants, clientID)
	name := p.Name
	isTeacher := p.Client.Role == message.RoleTeacher
	r.mu.Unlock()

	r.logger.Info("участник отключился", "client_id", clientID, "name", name)

	if isTeacher {
		if r.GetStatus() == StatusActive {
			_ = r.FinishSession(clientID, true)
		}
		return
	}

	r.broadcast(message.New(message.TypeParticipantLeft, message.ParticipantLeftPayload{
		Name:       name,
		TotalCount: r.StudentCount(),
	}))
}

// HandleMessage диспетчеризует входящее сообщение от клиента.
func (r *Room) HandleMessage(c *client.Client, msg message.IncomingMessage) {
	switch msg.Type {
	case message.TypePing:
		c.SendMsg(message.Pong())

	case message.TypeStartSession:
		if err := r.StartSession(c.ID); err != nil {
			c.SendMsg(message.NewError(message.ErrCodeUnauthorized, err.Error()))
		}

	case message.TypeNextQuestion:
		if err := r.NextQuestion(c.ID); err != nil {
			c.SendMsg(message.NewError(message.ErrCodeUnauthorized, err.Error()))
		}

	case message.TypeFinishSession:
		if err := r.FinishSession(c.ID, false); err != nil {
			c.SendMsg(message.NewError(message.ErrCodeUnauthorized, err.Error()))
		}

	case message.TypeAnswer:
		if err := r.SubmitAnswer(c, msg.QuestionID, msg.AnswerIndex); err != nil {
			c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, err.Error()))
		}

	default:
		c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, "неизвестный тип сообщения: "+msg.Type))
	}
}

// StartSession запускает квиз. Только преподаватель может запустить сессию.
func (r *Room) StartSession(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if clientID != r.TeacherID {
		return errors.New("только преподаватель может запустить сессию")
	}
	if r.Status != StatusWaiting {
		return errors.New("сессия уже запущена или завершена")
	}
	if len(r.Questions) == 0 {
		return errors.New("в квизе нет вопросов")
	}

	r.Status = StatusActive
	r.StartedAt = time.Now()
	r.CurrentQuestionIndex = -1

	r.logger.Info("сессия запущена", "quiz_title", r.QuizTitle, "participants", len(r.Participants))

	r.broadcastLocked(message.New(message.TypeSessionStarted, message.SessionStartedPayload{
		QuizTitle:      r.QuizTitle,
		TotalQuestions: len(r.Questions),
	}))

	r.nextQuestionLocked()
	return nil
}

// NextQuestion переходит к следующему вопросу.
func (r *Room) NextQuestion(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if clientID != r.TeacherID {
		return errors.New("только преподаватель может переключать вопросы")
	}
	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}

	return r.nextQuestionLocked()
}

// nextQuestionLocked — внутренний переход к следующему вопросу.
func (r *Room) nextQuestionLocked() error {
	if r.CurrentQuestionIndex >= 0 {
		r.sendQuestionResultsLocked()
	}

	r.CurrentQuestionIndex++

	if r.CurrentQuestionIndex >= len(r.Questions) {
		// Все вопросы пройдены — завершаем сессию
		r.finishLocked()
		return nil
	}

	q := r.Questions[r.CurrentQuestionIndex]
	timeLimitSec := q.TimeLimitSec
	if timeLimitSec <= 0 {
		timeLimitSec = r.cfg.DefaultQuestionTimeSec
	}

	r.QuestionStartedAt = time.Now()

	opts := make([]message.QuestionOption, len(q.Options))
	for i, o := range q.Options {
		opts[i] = message.QuestionOption{Index: i, Text: o}
	}

	r.broadcastLocked(message.New(message.TypeQuestion, message.QuestionPayload{
		QuestionID:   q.ID,
		Index:        r.CurrentQuestionIndex + 1,
		Total:        len(r.Questions),
		Text:         q.Text,
		Options:      opts,
		TimeLimitSec: timeLimitSec,
		StartedAt:    r.QuestionStartedAt,
	}))

	r.logger.Info("отправлен вопрос",
		"index", r.CurrentQuestionIndex+1,
		"total", len(r.Questions),
		"question_id", q.ID,
	)
	return nil
}

// SubmitAnswer принимает ответ студента.
func (r *Room) SubmitAnswer(c *client.Client, questionID string, answerIndex int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}
	if r.CurrentQuestionIndex < 0 || r.CurrentQuestionIndex >= len(r.Questions) {
		return errors.New("вопрос не найден")
	}

	currentQ := r.Questions[r.CurrentQuestionIndex]
	if currentQ.ID != questionID {
		return errors.New("вопрос уже сменился")
	}

	p, ok := r.Participants[c.ID]
	if !ok {
		return errors.New("участник не найден в комнате")
	}
	if _, answered := p.Answers[questionID]; answered {
		return errors.New("ответ уже принят")
	}

	isCorrect := answerIndex == currentQ.CorrectIndex
	score := r.calcScore(isCorrect, r.QuestionStartedAt, currentQ.TimeLimitSec)

	p.Answers[questionID] = Answer{
		AnswerIndex: answerIndex,
		AnsweredAt:  time.Now(),
		IsCorrect:   isCorrect,
		Score:       score,
	}
	if isCorrect {
		p.TotalScore += score
	}

	c.SendMsg(message.New(message.TypeAnswerResult, message.AnswerResultPayload{
		Correct:      isCorrect,
		CorrectIndex: currentQ.CorrectIndex,
		Score:        score,
		TotalScore:   p.TotalScore,
	}))

	answeredCount := r.answersCountLocked(questionID)
	if teacher := r.teacherClientLocked(); teacher != nil {
		teacher.SendMsg(message.New(message.TypeAnswerReceived, message.AnswerReceivedPayload{
			ParticipantName:   p.Name,
			AnswersCount:      answeredCount,
			TotalParticipants: len(r.students()),
		}))
	}

	r.logger.Debug("принят ответ",
		"participant", p.Name,
		"question_id", questionID,
		"correct", isCorrect,
		"score", score,
	)
	return nil
}

// FinishSession завершает сессию и формирует итоги.
func (r *Room) FinishSession(clientID string, force bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !force && clientID != r.TeacherID {
		return errors.New("только преподаватель может завершить сессию")
	}
	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}

	if r.CurrentQuestionIndex >= 0 {
		r.sendQuestionResultsLocked()
	}

	r.finishLocked()
	return nil
}

// finishLocked завершает сессию (вызывается под мьютексом).
func (r *Room) finishLocked() {
	r.Status = StatusFinished
	r.FinishedAt = time.Now()
	duration := int(r.FinishedAt.Sub(r.StartedAt).Seconds())

	results := r.buildResultsLocked()
	payload := message.SessionFinishedPayload{
		QuizTitle: r.QuizTitle,
		Results:   results,
		Duration:  duration,
	}

	r.broadcastLocked(message.New(message.TypeSessionFinished, payload))
	r.logger.Info("сессия завершена", "duration_sec", duration, "participants", len(r.Participants))

	if r.onFinish != nil {
		go r.onFinish(r)
	}
}

// sendQuestionResultsLocked рассылает итоги по текущему вопросу.
func (r *Room) sendQuestionResultsLocked() {
	if r.CurrentQuestionIndex < 0 || r.CurrentQuestionIndex >= len(r.Questions) {
		return
	}
	q := r.Questions[r.CurrentQuestionIndex]

	counts := make([]int, len(q.Options))
	for _, p := range r.Participants {
		if ans, ok := p.Answers[q.ID]; ok {
			if ans.AnswerIndex >= 0 && ans.AnswerIndex < len(counts) {
				counts[ans.AnswerIndex]++
			}
		}
	}
	stats := make([]message.AnswerStat, len(counts))
	for i, c := range counts {
		stats[i] = message.AnswerStat{OptionIndex: i, Count: c}
	}

	r.broadcastLocked(message.New(message.TypeQuestionResults, message.QuestionResultsPayload{
		QuestionID:   q.ID,
		CorrectIndex: q.CorrectIndex,
		Stats:        stats,
		Leaderboard:  r.buildLeaderboardLocked(5),
	}))
}

// calcScore вычисляет очки за ответ с учётом скорости.
func (r *Room) calcScore(correct bool, questionStart time.Time, timeLimitSec int) int {
	if !correct {
		return 0
	}
	elapsed := time.Since(questionStart).Seconds()
	limit := float64(timeLimitSec)
	if limit <= 0 {
		limit = float64(r.cfg.DefaultQuestionTimeSec)
	}
	if elapsed >= limit {
		return 100 // минимальные очки за правильный ответ
	}
	ratio := elapsed / limit
	score := int(1000 - 900*ratio)
	if score < 100 {
		score = 100
	}
	return score
}

// buildLeaderboardLocked строит топ-N участников.
func (r *Room) buildLeaderboardLocked(topN int) []message.ScoreEntry {
	type entry struct {
		name  string
		score int
	}
	var entries []entry
	for _, p := range r.Participants {
		if p.Client.Role == message.RoleStudent {
			entries = append(entries, entry{p.Name, p.TotalScore})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score > entries[j].score
	})
	if topN > 0 && len(entries) > topN {
		entries = entries[:topN]
	}
	result := make([]message.ScoreEntry, len(entries))
	for i, e := range entries {
		result[i] = message.ScoreEntry{Rank: i + 1, Name: e.name, Score: e.score}
	}
	return result
}

// buildResultsLocked формирует итоговую таблицу результатов.
func (r *Room) buildResultsLocked() []message.ParticipantResult {
	var results []message.ParticipantResult
	for _, p := range r.Participants {
		if p.Client.Role != message.RoleStudent {
			continue
		}
		correct := 0
		for _, a := range p.Answers {
			if a.IsCorrect {
				correct++
			}
		}
		results = append(results, message.ParticipantResult{
			Name:           p.Name,
			Score:          p.TotalScore,
			CorrectAnswers: correct,
			TotalQuestions: len(r.Questions),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

// broadcast рассылает сообщение всем участникам комнаты.
func (r *Room) broadcast(msg message.OutgoingMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.broadcastLocked(msg)
}

// broadcastLocked — broadcast без захвата мьютекса (вызывается под блокировкой).
func (r *Room) broadcastLocked(msg message.OutgoingMessage) {
	for _, p := range r.Participants {
		p.Client.SendMsg(msg)
	}
}

// broadcastToStudentsLocked рассылает только студентам.
func (r *Room) broadcastToStudentsLocked(msg message.OutgoingMessage) {
	for _, p := range r.Participants {
		if p.Client.Role == message.RoleStudent {
			p.Client.SendMsg(msg)
		}
	}
}

// teacherClientLocked возвращает клиента-преподавателя.
func (r *Room) teacherClientLocked() *client.Client {
	if p, ok := r.Participants[r.TeacherID]; ok {
		return p.Client
	}
	return nil
}

// students возвращает список студентов (без мьютекса — вызывать под r.mu).
func (r *Room) students() []*Participant {
	var s []*Participant
	for _, p := range r.Participants {
		if p.Client.Role == message.RoleStudent {
			s = append(s, p)
		}
	}
	return s
}

// answersCountLocked подсчитывает количество ответов на вопрос.
func (r *Room) answersCountLocked(questionID string) int {
	count := 0
	for _, p := range r.Participants {
		if _, ok := p.Answers[questionID]; ok {
			count++
		}
	}
	return count
}

// GetStatus возвращает текущий статус комнаты.
func (r *Room) GetStatus() Status {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Status
}

// StudentCount возвращает количество студентов.
func (r *Room) StudentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.students())
}

// GetParticipants возвращает снимок списка участников (для REST API).
func (r *Room) GetParticipants() []message.ParticipantInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []message.ParticipantInfo
	for _, p := range r.Participants {
		result = append(result, message.ParticipantInfo{
			Name: p.Name,
			ID:   p.Client.ID,
		})
	}
	return result
}

// GetLeaderboard возвращает топ участников.
func (r *Room) GetLeaderboard(topN int) []message.ScoreEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.buildLeaderboardLocked(topN)
}

// GetResults возвращает итоговые результаты (после завершения сессии).
func (r *Room) GetResults() []message.ParticipantResult {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.buildResultsLocked()
}

// IsFinished возвращает true, если сессия завершена.
func (r *Room) IsFinished() bool {
	return r.GetStatus() == StatusFinished
}
