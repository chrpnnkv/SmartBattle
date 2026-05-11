package handlers

// Детерминированные хелперы для аватарок участников. Используются ReportHandler
// при сборке leaderboard-DTO. Тот же алгоритм продублирован на фронте и в realtime —
// при желании выносится в отдельный пакет, но сейчас три копии живут отдельно
// просто чтобы не трогать сетевые границы.

func avatarInitials(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return "?"
	}
	parts := []rune{}
	prevSpace := true
	for _, r := range runes {
		if r == ' ' || r == '\t' {
			prevSpace = true
			continue
		}
		if prevSpace {
			parts = append(parts, r)
			if len(parts) == 2 {
				break
			}
		}
		prevSpace = false
	}
	if len(parts) == 0 {
		parts = append(parts, runes[0])
	}
	return string(parts)
}

var avatarPalette = []string{
	"#7c3aed", "#2563eb", "#16a34a", "#dc2626",
	"#ea580c", "#0891b2", "#be185d", "#d97706",
}

func avatarColor(name string) string {
	hash := 0
	for _, ch := range name {
		hash = int(ch) + ((hash << 5) - hash)
	}
	if hash < 0 {
		hash = -hash
	}
	return avatarPalette[hash%len(avatarPalette)]
}
