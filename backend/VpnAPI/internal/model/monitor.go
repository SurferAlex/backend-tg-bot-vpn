package model

import "github.com/gofrs/uuid/v5"

type MonitorTarget struct {
	ClientUUID     uuid.UUID
	MaxIPs         int
	XUIClientEmail string
}
