package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"quiz-core/internal/config"
	"quiz-core/internal/models"
	"quiz-core/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	services *service.Service
	cfg      *config.Config
}

func NewHandler(services *service.Service, cfg *config.Config) *Handler {
	return &Handler{services: services, cfg: cfg}
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создает аккаунт преподавателя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body object{email=string,password=string} true "Данные регистрации"
// @Success      201  "Created"
// @Failure      400  {string} string "Bad Request"
// @Failure      500  {string} string "Internal Server Error"
// @Router       /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.services.RegisterUser(r.Context(), req.Email, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Login godoc
// @Summary      Вход в систему
// @Description  Возвращает JWT токен
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body object{email=string,password=string} true "Данные входа"
// @Success      200  {object} map[string]string{token=string}
// @Failure      400  {string} string "Bad Request"
// @Failure      401  {string} string "Unauthorized"
// @Router       /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.services.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// GetMe godoc
// @Summary      Профиль пользователя
// @Description  Возвращает данные текущего авторизованного пользователя
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} models.User
// @Failure      401  {string} string "Unauthorized"
// @Failure      404  {string} string "Not Found"
// @Router       /api/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	user, err := h.services.GetUserProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// ChangePassword godoc
// @Summary      Смена пароля
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Param        input body object{old_password=string,new_password=string} true "Смена пароля"
// @Success      200  "OK"
// @Failure      400  {string} string "Error"
// @Router       /auth/change-password [post]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.services.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ForgotPassword godoc
// @Summary      Запрос на сброс пароля
// @Tags         Auth
// @Accept       json
// @Param        input body object{email=string} true "Email"
// @Success      200  {string} string "Instructions sent"
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_ = h.services.ForgotPassword(r.Context(), req.Email)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("If account exists, reset link sent to console"))
}

// ResetPassword godoc
// @Summary      Установка нового пароля по токену
// @Tags         Auth
// @Accept       json
// @Param        input body object{token=string,new_password=string} true "Reset Data"
// @Success      200  "OK"
// @Failure      400  {string} string "Error"
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.services.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// CreateQuiz godoc
// @Summary      Создание Квиза
// @Description  Создает квиз с вопросами и вариантами ответов
// @Tags         Quizzes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input body models.Quiz true "Структура квиза"
// @Success      200  {object} models.Quiz
// @Failure      400  {string} string "Validation Error"
// @Router       /api/quizzes [post]
func (h *Handler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	var quiz models.Quiz
	if err := json.NewDecoder(r.Body).Decode(&quiz); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	quiz.TeacherID = userID
	quiz.CreatedAt = time.Now()

	if err := h.services.CreateQuiz(r.Context(), &quiz); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(quiz)
}

// UpdateQuiz godoc
// @Summary      Обновление Квиза
// @Description  Полностью заменяет вопросы квиза на новые
// @Tags         Quizzes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Quiz ID"
// @Param        input body models.Quiz true "Новая структура"
// @Success      200  {object} models.Quiz
// @Failure      400  {string} string "Error"
// @Router       /api/quizzes/{id} [put]
func (h *Handler) UpdateQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var quiz models.Quiz
	if err := json.NewDecoder(r.Body).Decode(&quiz); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	quiz.ID = id
	quiz.TeacherID = userID

	if err := h.services.UpdateQuiz(r.Context(), &quiz); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(quiz)
}

// ListQuizzes godoc
// @Summary      Список квизов преподавателя
// @Tags         Quizzes
// @Security     BearerAuth
// @Param        page query int false "Page number"
// @Param        limit query int false "Items per page"
// @Produce      json
// @Success      200  {array} models.Quiz
// @Router       /api/quizzes [get]
func (h *Handler) ListQuizzes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	quizzes, err := h.services.GetTeacherQuizzes(r.Context(), userID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(quizzes)
}

// ListPublicQuizzes godoc
// @Summary      Публичная библиотека квизов
// @Tags         Quizzes
// @Security     BearerAuth
// @Param        page query int false "Page number"
// @Param        limit query int false "Items per page"
// @Produce      json
// @Success      200  {array} models.Quiz
// @Router       /api/quizzes/public [get]
func (h *Handler) ListPublicQuizzes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	quizzes, err := h.services.GetPublicQuizzes(r.Context(), page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(quizzes)
}

// GetQuiz godoc
// @Summary      Получить квиз по ID
// @Tags         Quizzes
// @Security     BearerAuth
// @Param        id   path int true "Quiz ID"
// @Produce      json
// @Success      200  {object} models.Quiz
// @Router       /api/quizzes/{id} [get]
func (h *Handler) GetQuiz(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	quiz, err := h.services.GetQuizFull(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(quiz)
}

// DeleteQuiz godoc
// @Summary      Удалить квиз
// @Tags         Quizzes
// @Security     BearerAuth
// @Param        id   path int true "Quiz ID"
// @Success      200  "OK"
// @Router       /api/quizzes/{id} [delete]
func (h *Handler) DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.services.DeleteQuiz(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// InternalGetQuiz godoc
// @Summary      INTERNAL: Получить квиз (для движка)
// @Tags         Internal
// @Param        id   path int true "Quiz ID"
// @Param        X-Internal-Secret header string true "Secret Key"
// @Produce      json
// @Success      200  {object} models.Quiz
// @Router       /internal/quizzes/{id} [get]
func (h *Handler) InternalGetQuiz(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Internal-Secret")
	if secret != h.cfg.InternalSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	quiz, err := h.services.GetQuizFull(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(quiz)
}

// InternalSaveReport godoc
// @Summary      INTERNAL: Сохранить результаты игры
// @Tags         Internal
// @Param        X-Internal-Secret header string true "Secret Key"
// @Param        input body models.GameSession true "Snapshot"
// @Success      201  "Created"
// @Router       /internal/reports [post]
func (h *Handler) InternalSaveReport(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Internal-Secret")
	if secret != h.cfg.InternalSecret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var session models.GameSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.services.SaveSessionReport(r.Context(), &session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ListReports godoc
// @Summary      История игр преподавателя
// @Tags         Analytics
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array} models.GameSession
// @Router       /api/reports [get]
func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uuid.UUID)
	reports, err := h.services.GetReports(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(reports)
}

// ExportReportCSV godoc
// @Summary      Скачать CSV отчет
// @Tags         Analytics
// @Security     BearerAuth
// @Param        id   path string true "Session UUID"
// @Produce      text/csv
// @Success      200  {file} file
// @Router       /api/reports/{id}/export [get]
func (h *Handler) ExportReportCSV(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, err := h.services.GetReportExportData(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var snapshot models.ReportSnapshot
	if err := json.Unmarshal(session.ReportSnapshot, &snapshot); err != nil {
		fmt.Printf("Error unmarshaling snapshot for session %s: %v\n", id, err)
	}

	filename := fmt.Sprintf("report_quiz%d_%s.csv", session.QuizID, session.FinishedAt.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"Session Report"})
	writer.Write([]string{"Session ID", session.ID.String()})
	writer.Write([]string{"Quiz ID", fmt.Sprintf("%d", session.QuizID)})
	writer.Write([]string{"Date", session.FinishedAt.Format(time.RFC1123)})
	writer.Write([]string{""})

	header := []string{"Rank", "Student Name", "Total Score"}
	for i := range snapshot.QuestionIDs {
		header = append(header, fmt.Sprintf("Q%d", i+1))
	}
	writer.Write(header)

	for _, p := range snapshot.Participants {
		row := []string{
			strconv.Itoa(p.Rank),
			p.Name,
			strconv.Itoa(p.Score),
		}
		for _, qID := range snapshot.QuestionIDs {
			if isCorrect, ok := p.Answers[qID]; ok {
				if isCorrect {
					row = append(row, "+")
				} else {
					row = append(row, "-")
				}
			} else {
				row = append(row, "")
			}
		}
		writer.Write(row)
	}
}
