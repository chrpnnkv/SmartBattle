package repository

import (
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

func (r *QuizRepository) Create(quiz *models.Quiz) error {
	return r.db.Create(quiz).Error
}

func (r *QuizRepository) GetByID(id uuid.UUID) (*models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.Preload("Questions.Options").First(&quiz, "id = ?", id).Error
	return &quiz, err
}

func (r *QuizRepository) GetByTeacherID(teacherID uuid.UUID) ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.Preload("Questions.Options").Where("teacher_id = ?", teacherID).Find(&quizzes).Error
	return quizzes, err
}

func (r *QuizRepository) GetAll() ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.Preload("Questions.Options").
		Order("created_at desc").
		Find(&quizzes).Error
	return quizzes, err
}

func (r *QuizRepository) GetPublic() ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.Preload("Questions.Options").
		Where("status = ?", "published").
		Order("created_at desc").
		Find(&quizzes).Error
	return quizzes, err
}

func (r *QuizRepository) Update(quiz *models.Quiz) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("quiz_id = ?", quiz.ID).Delete(&models.Question{}).Error; err != nil {
			return err
		}
		if err := tx.Save(quiz).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *QuizRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Quiz{}, "id = ?", id).Error
}
