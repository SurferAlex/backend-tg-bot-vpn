package handlers

import (
	"api-vpn/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	DB        *pgxpool.Pool
	Clients   *usecase.VPNClients
	XUIAccess *usecase.XUIAccess
}
