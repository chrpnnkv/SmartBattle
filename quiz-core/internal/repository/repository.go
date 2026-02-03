package repository

import (
	"context"
	"quiz-core/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

// Установка токена сброса
func (r *Repository) SetResetToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error
}

// Поиск пользователя по токену (и проверка срока действия)
func (r *Repository) GetUserByResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&user).Error
	return &user, err
}

// Очистка токена после успешного сброса + смена пароля
func (r *Repository) ResetPassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_hash":      newHash,
		"reset_token":        nil,
		"reset_token_expiry": nil,
	}).Error
}

func (r *Repository) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	for i := range quiz.Questions {
		if quiz.Questions[i].ID == uuid.Nil {
			quiz.Questions[i].ID = uuid.New()
		}
		for j := range quiz.Questions[i].Options {
			if quiz.Questions[i].Options[j].ID == "" {
				quiz.Questions[i].Options[j].ID = uuid.New().String()
			}
			quiz.Questions[i].Options[j].QuestionID = quiz.Questions[i].ID
		}
	}
	return r.db.WithContext(ctx).Create(quiz).Error
}

func (r *Repository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Quiz{}).Where("id = ?", quiz.ID).Updates(map[string]interface{}{
			"title":       quiz.Title,
			"description": quiz.Description,
			"settings":    quiz.Settings,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("quiz_id = ?", quiz.ID).Delete(&models.Question{}).Error; err != nil {
			return err
		}
		for i := range quiz.Questions {
			quiz.Questions[i].QuizID = quiz.ID
			if quiz.Questions[i].ID == uuid.Nil {
				quiz.Questions[i].ID = uuid.New()
			}
			for j := range quiz.Questions[i].Options {
				if quiz.Questions[i].Options[j].ID == "" {
					quiz.Questions[i].Options[j].ID = uuid.New().String()
				}
				quiz.Questions[i].Options[j].QuestionID = quiz.Questions[i].ID
			}
		}
		if len(quiz.Questions) > 0 {
			if err := tx.Create(&quiz.Questions).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetQuizzesByTeacher(ctx context.Context, teacherID uuid.UUID, limit, offset int) ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Limit(limit).Offset(offset).
		Order("created_at desc").
		Find(&quizzes).Error
	return quizzes, err
}

func (r *Repository) GetPublicQuizzes(ctx context.Context, limit, offset int) ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.WithContext(ctx).
		Limit(limit).Offset(offset).
		Order("created_at desc").
		Find(&quizzes).Error
	return quizzes, err
}

func (r *Repository) GetFullQuiz(ctx context.Context, id int64) (*models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.WithContext(ctx).
		Preload("Questions.Options").
		First(&quiz, id).Error
	return &quiz, err
}

func (r *Repository) DeleteQuiz(ctx context.Context, id int64, teacherID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND teacher_id = ?", id, teacherID).
		Delete(&models.Quiz{}).Error
}

func (r *Repository) SaveGameSession(ctx context.Context, session *models.GameSession) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *Repository) GetSessionsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.WithContext(ctx).
		Where("host_id = ?", teacherID).
		Order("finished_at desc").
		Find(&sessions).Error
	return sessions, err
}

func (r *Repository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.WithContext(ctx).First(&session, id).Error
	return &session, err
}
