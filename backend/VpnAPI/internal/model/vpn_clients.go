package model

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type VPNClient struct {
	ID             int64
	ClientUUID     uuid.UUID
	TelegramUserID *int64
	MaxIPs         int
	KeyExpiresAt   time.Time
	IsActive       bool
	Note           *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateVPNClientParams struct {
	ClientUUID     uuid.UUID
	TelegramUserID *int64
	MaxIPs         int
	KeyExpiresAt   time.Time
	Note           *string
}
