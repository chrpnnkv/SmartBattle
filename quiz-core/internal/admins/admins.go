// Package admins реализует загрузку списка администраторов из статического JSON-файла.
//
// Решение принципиально простое: вместо отдельной таблицы users-with-role-admin
// и интерфейса управления учётными записями, права администратора выдаются
// через перечисление email-адресов в файле admins.json. Файл редактируется
// обслуживающим персоналом и перечитывается при перезапуске сервиса.
//
// На уровне кода: AuthService при выпуске JWT-токена проверяет email
// пользователя через метод IsAdmin; если адрес присутствует в списке —
// роль в полезной нагрузке токена устанавливается в "admin", независимо
// от того, что записано в таблице users. Обработчики, в свою очередь,
// читают роль из контекста Gin (положена туда middleware AuthGuard) и при
// значении "admin" расширяют выборку до всех квизов/отчётов.
package admins

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
)

// Role admin задаётся как константа, чтобы не было опечаток в обработчиках.
const RoleAdmin = "admin"

// List — потокобезопасный список email'ов администраторов.
// Поиск регистронезависимый и устойчив к пробелам.
type List struct {
	mu     sync.RWMutex
	emails map[string]struct{}
}

// fileSchema — структура содержимого admins.json.
//
// Пример файла:
//
//	{
//	    "admins": [
//	        "admin@hse.ru",
//	        "lecturer-coordinator@university.org"
//	    ]
//	}
type fileSchema struct {
	Admins []string `json:"admins"`
}

// New создаёт пустой список (полезно для тестов, где admins.json не нужен).
func New() *List {
	return &List{emails: map[string]struct{}{}}
}

// LoadFromFile читает admins.json и возвращает заполненный List.
// Если файл отсутствует, возвращает пустой List и пишет предупреждение в лог:
// в учебных сценариях это допустимый режим работы (никто не считается администратором).
func LoadFromFile(path string) (*List, error) {
	l := New()
	if path == "" {
		log.Printf("admins: путь к файлу не задан — список администраторов пуст")
		return l, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("admins: файл %s не найден — список администраторов пуст", path)
			return l, nil
		}
		return nil, err
	}

	var parsed fileSchema
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	for _, email := range parsed.Admins {
		normalized := normalize(email)
		if normalized != "" {
			l.emails[normalized] = struct{}{}
		}
	}
	log.Printf("admins: загружено %d администратор(ов) из %s", len(l.emails), path)
	return l, nil
}

// IsAdmin возвращает true, если email присутствует в списке.
// Сравнение регистронезависимое; ведущие и завершающие пробелы игнорируются.
func (l *List) IsAdmin(email string) bool {
	if l == nil {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.emails[normalize(email)]
	return ok
}

// Count возвращает число администраторов; используется для логирования.
func (l *List) Count() int {
	if l == nil {
		return 0
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.emails)
}

func normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
