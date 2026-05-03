package service

import "testing"

// sanitizePIN — приватная функция в этом пакете. Тест проверяет, что любые
// варианты ввода с фронта/реалтайма (с пробелами, дефисами, в нижнем регистре,
// смешанные) приводятся к одной канонической форме, которая уходит в БД и
// используется для GetByPIN. Если эта функция сломается, ни одна сессия
// не сматчится между FE → Core → Realtime.
func TestSanitizePIN(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain digits", "483480", "483480"},
		{"with single space in middle", "483 480", "483480"},
		{"with multiple spaces", "  483  480  ", "483480"},
		{"with dash", "483-480", "483480"},
		{"with dash and spaces", " 483 - 480 ", "483480"},
		{"lowercase letters", "abcd12", "ABCD12"},
		{"mixed case with separators", "ab cd-12", "ABCD12"},
		{"empty", "", ""},
		{"only separators", "  - -  ", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizePIN(tc.in)
			if got != tc.want {
				t.Fatalf("sanitizePIN(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Идемпотентность: пропустить уже очищенный PIN через sanitizePIN ещё раз
// должно возвращать тот же результат. Это важно потому, что Realtime пишет PIN
// в snapshot уже сырым, а Core снова применяет sanitizePIN при поиске сессии.
func TestSanitizePINIdempotent(t *testing.T) {
	cleaned := sanitizePIN("ab cd-12")
	again := sanitizePIN(cleaned)
	if cleaned != again {
		t.Fatalf("sanitizePIN не идемпотентна: %q -> %q -> %q", "ab cd-12", cleaned, again)
	}
}
