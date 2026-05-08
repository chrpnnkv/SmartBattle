package admins

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
)

const RoleAdmin = "admin"

type List struct {
	mu     sync.RWMutex
	emails map[string]struct{}
}
type fileSchema struct {
	Admins []string `json:"admins"`
}

func New() *List {
	return &List{emails: map[string]struct{}{}}
}

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

func (l *List) IsAdmin(email string) bool {
	if l == nil {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.emails[normalize(email)]
	return ok
}

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
