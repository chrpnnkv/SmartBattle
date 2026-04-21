package room

import (
	"errors"
	"log/slog"
	"sort"
	"strings"
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

// OptionColors — цвета вариантов ответа (по порядку).
var OptionColors = []string{"red", "blue", "yellow", "green"}

// QuestionOption — вариант ответа вопроса (внутреннее представление).
type QuestionOption struct {
	ID        string
	Text      string
	IsCorrect bool
	Color     string
}

// Question — вопрос квиза (внутреннее представление).
type Question struct {
	ID           string
	QuizID       string
	Type         string // "multiple_choice", "true_false", etc.
	Text         string
	Options      []QuestionOption
	TimeLimitSec int
	Order        int
}

// OptionByID возвращает вариант ответа по его ID.
func (q *Question) OptionByID(id string) (QuestionOption, bool) {
	for _, o := range q.Options {
		if o.ID == id {
			return o, true
		}
	}
	return QuestionOption{}, false
}

// CorrectIndex возвращает индекс первого правильного варианта.
func (q *Question) CorrectIndex() int {
	for i, o := range q.Options {
		if o.IsCorrect {
			return i
		}
	}
	return 0
}

// Answer — ответ участника.
type Answer struct {
	OptionID   string
	AnsweredAt time.Time
	IsCorrect  bool
	Score      int
	ResponseMs int64 // время от старта вопроса до ответа в мс
}

// Participant — участник сессии.
type Participant struct {
	Client         *client.Client
	ParticipantID  string // стабильный UUID (из предрегистрации или client.ID)
	Name           string
	AvatarInitials string
	AvatarColor    string
	TotalScore     int
	AnsweredCount  int
	Answers        map[string]Answer
}

// Room — игровая комната.
type Room struct {
	mu   sync.RWMutex
	Code string

	Status    Status
	QuizID    string
	QuizTitle string
	Mode      string
	Questions []Question

	TeacherID     string // ID WS-клиента преподавателя
	TeacherUserID string // UserID из JWT (для REST-контроля)
	Participants  map[string]*Participant

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

// isTeacher проверяет, является ли caller преподавателем этой комнаты.
func (r *Room) isTeacher(callerID string) bool {
	return callerID == r.TeacherID ||
		(r.TeacherUserID != "" && callerID == r.TeacherUserID)
}

// AddClient добавляет клиента в комнату. participantID — стабильный UUID участника.
func (r *Room) AddClient(c *client.Client, participantID string) error {
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

	initials := avatarInitials(c.Name)
	color := avatarColor(c.Name)

	r.Participants[c.ID] = &Participant{
		Client:         c,
		ParticipantID:  participantID,
		Name:           c.Name,
		AvatarInitials: initials,
		AvatarColor:    color,
		Answers:        make(map[string]Answer),
	}

	r.logger.Info("участник подключился", "client_id", c.ID, "participant_id", participantID, "name", c.Name, "role", c.Role)
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
	participantID := p.ParticipantID
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
		ParticipantID: participantID,
		Name:          name,
		TotalCount:    r.StudentCount(),
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
		if err := r.submitAnswerFromWS(c, msg); err != nil {
			c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, err.Error()))
		}

	default:
		c.SendMsg(message.NewError(message.ErrCodeInvalidMessage, "неизвестный тип сообщения: "+msg.Type))
	}
}

// StartSession запускает квиз. Только преподаватель может запустить сессию.
func (r *Room) StartSession(callerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isTeacher(callerID) {
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
func (r *Room) NextQuestion(callerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isTeacher(callerID) {
		return errors.New("только преподаватель может переключать вопросы")
	}
	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}

	return r.nextQuestionLocked()
}

// nextQuestionLocked — внутренний переход к следующему вопросу (вызывается под мьютексом).
func (r *Room) nextQuestionLocked() error {
	if r.CurrentQuestionIndex >= 0 {
		r.sendQuestionEndedLocked()
	}

	r.CurrentQuestionIndex++

	if r.CurrentQuestionIndex >= len(r.Questions) {
		r.finishLocked()
		return nil
	}

	q := r.Questions[r.CurrentQuestionIndex]
	timeLimitSec := q.TimeLimitSec
	if timeLimitSec <= 0 {
		timeLimitSec = r.cfg.DefaultQuestionTimeSec
	}

	r.QuestionStartedAt = time.Now()

	opts := make([]message.AnswerOptionData, len(q.Options))
	for i, o := range q.Options {
		opts[i] = message.AnswerOptionData{
			ID:        o.ID,
			Text:      o.Text,
			IsCorrect: o.IsCorrect,
			Color:     o.Color,
		}
	}

	payload := message.QuestionStartedPayload{
		Question: message.QuestionData{
			ID:               q.ID,
			QuizID:           q.QuizID,
			Type:             q.Type,
			Text:             q.Text,
			Options:          opts,
			TimeLimitSeconds: timeLimitSec,
			Order:            r.CurrentQuestionIndex,
		},
		QuestionIndex:  r.CurrentQuestionIndex,
		TotalQuestions: len(r.Questions),
		StartedAt:      r.QuestionStartedAt.UnixMilli(),
	}

	r.broadcastLocked(message.New(message.TypeQuestion, payload))

	r.logger.Info("отправлен вопрос",
		"index", r.CurrentQuestionIndex+1,
		"total", len(r.Questions),
		"question_id", q.ID,
	)
	return nil
}

// submitAnswerFromWS обрабатывает ответ студента, пришедший по WebSocket.
func (r *Room) submitAnswerFromWS(c *client.Client, msg message.IncomingMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}
	if r.CurrentQuestionIndex < 0 || r.CurrentQuestionIndex >= len(r.Questions) {
		return errors.New("вопрос не найден")
	}

	currentQ := r.Questions[r.CurrentQuestionIndex]
	if msg.QuestionID != "" && currentQ.ID != msg.QuestionID {
		return errors.New("вопрос уже сменился")
	}

	p, ok := r.Participants[c.ID]
	if !ok {
		return errors.New("участник не найден в комнате")
	}
	if _, answered := p.Answers[currentQ.ID]; answered {
		return errors.New("ответ уже принят")
	}

	var opt QuestionOption
	var found bool

	// Предпочитаем AnswerID (string), если передан
	if msg.AnswerID != "" {
		opt, found = currentQ.OptionByID(msg.AnswerID)
	}
	// Иначе используем AnswerIndex
	if !found && msg.AnswerIndex >= 0 && msg.AnswerIndex < len(currentQ.Options) {
		opt = currentQ.Options[msg.AnswerIndex]
		found = true
	}
	if !found {
		return errors.New("некорректный вариант ответа")
	}

	responseMs := time.Since(r.QuestionStartedAt).Milliseconds()
	score := r.calcScore(opt.IsCorrect, r.QuestionStartedAt, currentQ.TimeLimitSec)

	p.Answers[currentQ.ID] = Answer{
		OptionID:   opt.ID,
		AnsweredAt: time.Now(),
		IsCorrect:  opt.IsCorrect,
		Score:      score,
		ResponseMs: responseMs,
	}
	p.AnsweredCount++
	if opt.IsCorrect {
		p.TotalScore += score
	}

	c.SendMsg(message.New(message.TypeAnswerResult, message.AnswerResultPayload{
		Correct:      opt.IsCorrect,
		CorrectIndex: currentQ.CorrectIndex(),
		Score:        score,
		TotalScore:   p.TotalScore,
	}))

	answeredCount := r.answersCountLocked(currentQ.ID)
	if teacher := r.teacherClientLocked(); teacher != nil {
		teacher.SendMsg(message.New(message.TypeAnswerReceived, message.AnswerReceivedPayload{
			ParticipantName:   p.Name,
			AnswersCount:      answeredCount,
			TotalParticipants: len(r.students()),
		}))
	}

	r.logger.Debug("принят ответ (WS)",
		"participant", p.Name,
		"question_id", currentQ.ID,
		"correct", opt.IsCorrect,
		"score", score,
	)
	return nil
}

// SubmitAnswerREST принимает ответ студента через REST API.
func (r *Room) SubmitAnswerREST(participantID, questionID, optionID string, responseMs int64) error {
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

	p := r.getParticipantByStableID(participantID)
	if p == nil {
		return errors.New("участник не найден в комнате")
	}
	if _, answered := p.Answers[questionID]; answered {
		c := p.Client
		if c != nil {
			c.SendMsg(message.NewError(message.ErrCodeAlreadyAnswered, "ответ уже принят"))
		}
		return errors.New("ответ уже принят")
	}

	opt, found := currentQ.OptionByID(optionID)
	if !found {
		return errors.New("вариант ответа не найден: " + optionID)
	}

	if responseMs <= 0 {
		responseMs = time.Since(r.QuestionStartedAt).Milliseconds()
	}
	score := r.calcScore(opt.IsCorrect, r.QuestionStartedAt, currentQ.TimeLimitSec)

	p.Answers[questionID] = Answer{
		OptionID:   opt.ID,
		AnsweredAt: time.Now(),
		IsCorrect:  opt.IsCorrect,
		Score:      score,
		ResponseMs: responseMs,
	}
	p.AnsweredCount++
	if opt.IsCorrect {
		p.TotalScore += score
	}

	// Отправляем результат студенту по WS, если он подключён
	if c := p.Client; c != nil {
		c.SendMsg(message.New(message.TypeAnswerResult, message.AnswerResultPayload{
			Correct:      opt.IsCorrect,
			CorrectIndex: currentQ.CorrectIndex(),
			Score:        score,
			TotalScore:   p.TotalScore,
		}))
	}

	answeredCount := r.answersCountLocked(questionID)
	if teacher := r.teacherClientLocked(); teacher != nil {
		teacher.SendMsg(message.New(message.TypeAnswerReceived, message.AnswerReceivedPayload{
			ParticipantName:   p.Name,
			AnswersCount:      answeredCount,
			TotalParticipants: len(r.students()),
		}))
	}

	r.logger.Debug("принят ответ (REST)",
		"participant", p.Name,
		"question_id", questionID,
		"correct", opt.IsCorrect,
		"score", score,
	)
	return nil
}

// FinishSession завершает сессию и формирует итоги.
func (r *Room) FinishSession(callerID string, force bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !force && !r.isTeacher(callerID) {
		return errors.New("только преподаватель может завершить сессию")
	}
	if r.Status != StatusActive {
		return errors.New("сессия не активна")
	}

	if r.CurrentQuestionIndex >= 0 {
		r.sendQuestionEndedLocked()
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

// sendQuestionEndedLocked рассылает итоги по текущему вопросу.
func (r *Room) sendQuestionEndedLocked() {
	if r.CurrentQuestionIndex < 0 || r.CurrentQuestionIndex >= len(r.Questions) {
		return
	}
	q := r.Questions[r.CurrentQuestionIndex]
	endedAt := time.Now()

	// Строим статистику распределения ответов
	distribution := make([]message.AnswerDistribution, len(q.Options))
	for i, opt := range q.Options {
		distribution[i] = message.AnswerDistribution{
			OptionID:   opt.ID,
			OptionText: opt.Text,
			Count:      0,
			IsCorrect:  opt.IsCorrect,
			Color:      opt.Color,
		}
	}

	var totalAnswers int
	var totalCorrect int
	var totalResponseMs int64
	var respondedCount int

	type correctAnswer struct {
		participantID string
		nickname      string
		responseMs    int64
	}
	var fastestCorrect []correctAnswer

	for _, p := range r.Participants {
		if p.Client.Role != message.RoleStudent {
			continue
		}
		ans, ok := p.Answers[q.ID]
		if !ok {
			continue
		}
		totalAnswers++
		totalResponseMs += ans.ResponseMs
		respondedCount++
		if ans.IsCorrect {
			totalCorrect++
			fastestCorrect = append(fastestCorrect, correctAnswer{
				participantID: p.ParticipantID,
				nickname:      p.Name,
				responseMs:    ans.ResponseMs,
			})
		}
		// Обновляем счётчик в distribution
		for i, opt := range q.Options {
			if opt.ID == ans.OptionID {
				distribution[i].Count++
				break
			}
		}
	}

	// Находим самый популярный неправильный ответ
	var mostWrongOptID, mostWrongOptText string
	var mostWrongCount int
	for _, d := range distribution {
		if !d.IsCorrect && d.Count > mostWrongCount {
			mostWrongCount = d.Count
			mostWrongOptID = d.OptionID
			mostWrongOptText = d.OptionText
		}
	}

	// Топ-5 самых быстрых правильных ответов
	sort.Slice(fastestCorrect, func(i, j int) bool {
		return fastestCorrect[i].responseMs < fastestCorrect[j].responseMs
	})
	if len(fastestCorrect) > 5 {
		fastestCorrect = fastestCorrect[:5]
	}
	fastestShort := make([]message.ParticipantShort, len(fastestCorrect))
	for i, fc := range fastestCorrect {
		fastestShort[i] = message.ParticipantShort{ID: fc.participantID, Nickname: fc.nickname}
	}

	correctPercent := 0
	if totalAnswers > 0 {
		correctPercent = int(float64(totalCorrect) / float64(totalAnswers) * 100)
	}
	avgResponseMs := 0
	if respondedCount > 0 {
		avgResponseMs = int(totalResponseMs / int64(respondedCount))
	}

	report := message.QuestionReport{
		QuestionID:                 q.ID,
		QuestionText:               q.Text,
		CorrectPercent:             correctPercent,
		AvgResponseTimeMs:          avgResponseMs,
		MostCommonWrongOptionID:    mostWrongOptID,
		MostCommonWrongOptionText:  mostWrongOptText,
		Distribution:               distribution,
		FastestCorrectParticipants: fastestShort,
	}

	payload := message.QuestionEndedPayload{
		QuestionReport: report,
		Leaderboard:    r.buildSessionParticipantsLocked(),
		EndedAt:        endedAt.UnixMilli(),
	}

	r.broadcastLocked(message.New(message.TypeQuestionResults, payload))

	entries := make([]message.ScoreEntry, len(payload.Leaderboard))
	for i, p := range payload.Leaderboard {
		entries[i] = message.ScoreEntry{Rank: i + 1, Name: p.Nickname, Score: p.Score}
	}
	r.broadcastLocked(message.New(message.TypeLeaderboard, message.LeaderboardPayload{Entries: entries}))
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
		return 100
	}
	ratio := elapsed / limit
	score := int(1000 - 900*ratio)
	if score < 100 {
		score = 100
	}
	return score
}

// buildSessionParticipantsLocked строит список участников с результатами (под мьютексом).
func (r *Room) buildSessionParticipantsLocked() []message.SessionParticipant {
	var participants []message.SessionParticipant
	for _, p := range r.Participants {
		if p.Client.Role == message.RoleStudent {
			participants = append(participants, message.SessionParticipant{
				ID:             p.ParticipantID,
				Nickname:       p.Name,
				AvatarInitials: p.AvatarInitials,
				AvatarColor:    p.AvatarColor,
				Score:          p.TotalScore,
				AnsweredCount:  p.AnsweredCount,
			})
		}
	}
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})
	return participants
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

// getParticipantByStableID ищет участника по его стабильному ParticipantID.
func (r *Room) getParticipantByStableID(participantID string) *Participant {
	for _, p := range r.Participants {
		if p.ParticipantID == participantID {
			return p
		}
	}
	return nil
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
func (r *Room) GetParticipants() []message.SessionParticipant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.buildSessionParticipantsLocked()
}

// GetLeaderboard возвращает топ участников.
func (r *Room) GetLeaderboard(topN int) []message.SessionParticipant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	participants := r.buildSessionParticipantsLocked()
	if topN > 0 && len(participants) > topN {
		return participants[:topN]
	}
	return participants
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

// avatarInitials вычисляет инициалы для аватара по имени.
func avatarInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	if len(parts) == 1 {
		r := []rune(parts[0])
		if len(r) >= 2 {
			return strings.ToUpper(string(r[:2]))
		}
		return strings.ToUpper(string(r))
	}
	r0 := []rune(parts[0])
	r1 := []rune(parts[1])
	if len(r0) == 0 || len(r1) == 0 {
		return "?"
	}
	return strings.ToUpper(string(r0[0])) + strings.ToUpper(string(r1[0]))
}

var avatarColorPalette = []string{
	"#7c3aed", "#2563eb", "#16a34a", "#dc2626",
	"#ea580c", "#0891b2", "#be185d", "#d97706",
}

// avatarColor вычисляет цвет аватара детерминированно по имени.
func avatarColor(name string) string {
	hash := 0
	for _, ch := range name {
		hash = int(ch) + ((hash << 5) - hash)
	}
	if hash < 0 {
		hash = -hash
	}
	return avatarColorPalette[hash%len(avatarColorPalette)]
}
