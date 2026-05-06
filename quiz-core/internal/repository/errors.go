package repository

import (
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound — package-local sentinel, чтобы сервисный слой не зависел напрямую
// от gorm. Вызывающий код использует errors.Is(err, repository.ErrNotFound).
var ErrNotFound = errors.New("repository: record not found")

// translate — переводит gorm.ErrRecordNotFound в наш sentinel, остальное —
// возвращает как есть. Используется внутри методов репозиториев.
func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
