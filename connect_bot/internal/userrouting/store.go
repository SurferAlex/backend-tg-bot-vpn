package userrouting

import (
	"strings"
	"sync"
)

// Store keeps the last Happ routing profile (base64) per Telegram chat.
type Store struct {
	mu sync.Map
}

func (s *Store) Set(chatID int64, profileB64 string) {
	profileB64 = strings.TrimSpace(profileB64)
	if profileB64 == "" {
		return
	}
	s.mu.Store(chatID, profileB64)
}

func (s *Store) Get(chatID int64) string {
	if v, ok := s.mu.Load(chatID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
