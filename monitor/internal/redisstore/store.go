package redisstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb       *redis.Client
	activeTTL time.Duration
	alertTTL  time.Duration
}

func New(rdb *redis.Client, activeTTL, alertTTL time.Duration) *Store {
	return &Store{rdb: rdb, activeTTL: activeTTL, alertTTL: alertTTL}
}

func activeKey(uuid, ip string) string {
	return fmt.Sprintf("active:%s:%s", uuid, ip)
}

func alertedKey(clientUUID, ip string) string {
	return fmt.Sprintf("alerted:%s:%s", clientUUID, ip)
}

// RecordConnection marks IP as active for the client (refreshes TTL).
func (s *Store) RecordConnection(ctx context.Context, clientUUID, ip string, at time.Time) error {
	return s.rdb.SetEx(ctx, activeKey(clientUUID, ip), at.Unix(), s.activeTTL).Err()
}

type ActiveIP struct {
	IP       string
	LastSeen time.Time
}

// ListActiveIPs returns IPs still within TTL for the client.
func (s *Store) ListActiveIPs(ctx context.Context, clientUUID string) ([]ActiveIP, error) {
	pattern := fmt.Sprintf("active:%s:*", clientUUID)
	var cursor uint64
	out := make([]ActiveIP, 0, 4)

	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			ip := strings.TrimPrefix(key, "active:"+clientUUID+":")
			if ip == "" || ip == key {
				continue
			}
			val, err := s.rdb.Get(ctx, key).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return nil, err
			}
			sec, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				continue
			}
			out = append(out, ActiveIP{IP: ip, LastSeen: time.Unix(sec, 0)})
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

func (s *Store) IsAlerted(ctx context.Context, clientUUID, ip string) (bool, error) {
	n, err := s.rdb.Exists(ctx, alertedKey(clientUUID, ip)).Result()
	return n > 0, err
}

func (s *Store) MarkAlerted(ctx context.Context, clientUUID, ip string) error {
	return s.rdb.SetEx(ctx, alertedKey(clientUUID, ip), "1", s.alertTTL).Err()
}

// ClearAlerted removes alert dedup for client+ip (for manual reset from bot later).
func (s *Store) ClearAlerted(ctx context.Context, clientUUID, ip string) error {
	return s.rdb.Del(ctx, alertedKey(clientUUID, ip)).Err()
}
