package room

import (
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/client"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
)

// BroadcastParticipantJoined уведомляет всех участников комнаты о входе нового студента.
func (r *Room) BroadcastParticipantJoined(newClient *client.Client, participantID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.Participants[newClient.ID]
	if !ok {
		return
	}

	studentCount := 0
	for _, part := range r.Participants {
		if part.Client.Role == message.RoleStudent {
			studentCount++
		}
	}

	msg := message.New(message.TypeParticipantJoined, message.ParticipantJoinedPayload{
		Participant: message.SessionParticipant{
			ID:             participantID,
			Nickname:       p.Name,
			AvatarInitials: p.AvatarInitials,
			AvatarColor:    p.AvatarColor,
			Score:          0,
			AnsweredCount:  0,
		},
		TotalCount: studentCount,
	})

	r.broadcastLocked(msg)
}

// SendLeaderboard рассылает текущий рейтинг всем участникам.
func (r *Room) SendLeaderboard() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	participants := r.buildSessionParticipantsLocked()
	entries := make([]message.ScoreEntry, len(participants))
	for i, p := range participants {
		entries[i] = message.ScoreEntry{
			Rank:  i + 1,
			Name:  p.Nickname,
			Score: p.Score,
		}
	}
	r.broadcastLocked(message.New(message.TypeLeaderboard, message.LeaderboardPayload{
		Entries: entries,
	}))
}
