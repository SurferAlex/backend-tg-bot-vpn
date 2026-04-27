package repository

import (
	"context"
	"errors"
	"time"

	"api-vpn/internal/model"
	"api-vpn/internal/usecase"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VPNClientsRepo struct {
	db *pgxpool.Pool
}

func NewVPNClientsRepo(db *pgxpool.Pool) *VPNClientsRepo {
	return &VPNClientsRepo{db: db}
}

func (r *VPNClientsRepo) GetByUUID(ctx context.Context, clientUUID uuid.UUID) (model.VPNClient, error) {
	const q = `
SELECT
  id, client_uuid, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at
FROM vpn_clients
WHERE client_uuid = $1
LIMIT 1;
`
	var c model.VPNClient
	err := r.db.QueryRow(ctx, q, clientUUID).Scan(
		&c.ID,
		&c.ClientUUID,
		&c.TelegramUserID,
		&c.MaxIPs,
		&c.KeyExpiresAt,
		&c.IsActive,
		&c.Note,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.VPNClient{}, usecase.ErrNotFound
	}
	return c, err
}

func (r *VPNClientsRepo) Create(ctx context.Context, p model.CreateVPNClientParams) (model.VPNClient, error) {
	const q = `
INSERT INTO vpn_clients (client_uuid, telegram_user_id, max_ips, key_expires_at, is_active, note)
VALUES ($1, $2, $3, $4, true, $5)
RETURNING id, client_uuid, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at;
`
	var c model.VPNClient
	err := r.db.QueryRow(ctx, q,
		p.ClientUUID,
		p.TelegramUserID,
		p.MaxIPs,
		p.KeyExpiresAt.UTC(),
		p.Note,
	).Scan(
		&c.ID,
		&c.ClientUUID,
		&c.TelegramUserID,
		&c.MaxIPs,
		&c.KeyExpiresAt,
		&c.IsActive,
		&c.Note,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	return c, err
}

func (r *VPNClientsRepo) Deactivate(ctx context.Context, clientUUID uuid.UUID) error {
	const q = `
UPDATE vpn_clients
SET is_active = false, updated_at = now()
WHERE client_uuid = $1 AND is_active = true;
`
	ct, err := r.db.Exec(ctx, q, clientUUID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return usecase.ErrNotFound
	}
	return nil
}

func (r *VPNClientsRepo) ExpireNow(ctx context.Context, clientUUID uuid.UUID) error {
	const q = `
UPDATE vpn_clients
SET key_expires_at = $2, updated_at = now()
WHERE client_uuid = $1;
`
	ct, err := r.db.Exec(ctx, q, clientUUID, time.Now().UTC())
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return usecase.ErrNotFound
	}
	return nil
}
