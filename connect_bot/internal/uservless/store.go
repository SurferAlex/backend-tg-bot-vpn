package uservless

import (
	"strings"
	"sync"
)

// Store keeps the last vless:// key per Telegram chat.
type Store struct {
	mu sync.Map
}

func (s *Store) Set(chatID int64, vless string) {
	vless = strings.TrimSpace(vless)
	if vless == "" || !strings.HasPrefix(strings.ToLower(vless), "vless://") {
		return
	}
	s.mu.Store(chatID, vless)
}

func (s *Store) Get(chatID int64) string {
	if v, ok := s.mu.Load(chatID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
