package model

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type XUIAccess struct {
	ID             int64
	ClientUUID     uuid.UUID
	InboundID      int64
	XUIClientEmail string
	VLESSURI       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UpsertXUIAccessParams struct {
	ClientUUID     uuid.UUID
	InboundID      int64
	XUIClientEmail string
	VLESSURI       string
}

