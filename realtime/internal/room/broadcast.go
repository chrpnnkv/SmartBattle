package room

import (
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/client"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
)

// BroadcastParticipantJoined уведомляет всех участников комнаты о входе нового студента.
func (r *Room) BroadcastParticipantJoined(newClient *client.Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]message.ParticipantInfo, 0, len(r.Participants))
	for _, p := range r.Participants {
		if p.Client.Role == message.RoleStudent {
			participants = append(participants, message.ParticipantInfo{
				Name: p.Name,
				ID:   p.Client.ID,
			})
		}
	}

	msg := message.New(message.TypeParticipantJoined, message.ParticipantJoinedPayload{
		Name:         newClient.Name,
		Participants: participants,
		TotalCount:   len(participants),
	})

	r.broadcastLocked(msg)
}

// SendLeaderboard рассылает текущий рейтинг всем участникам.
func (r *Room) SendLeaderboard() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.broadcastLocked(message.New(message.TypeLeaderboard, message.LeaderboardPayload{
		Entries: r.buildLeaderboardLocked(0),
	}))
}
