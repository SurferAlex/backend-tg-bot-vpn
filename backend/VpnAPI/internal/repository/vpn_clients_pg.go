package repository

import (
	"context"
	"errors"
	"strings"
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
  id, client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at
FROM vpn_clients
WHERE client_uuid = $1
LIMIT 1;
`
	var c model.VPNClient
	err := r.db.QueryRow(ctx, q, clientUUID).Scan(
		&c.ID,
		&c.ClientUUID,
		&c.ServerID,
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
INSERT INTO vpn_clients (client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note)
VALUES ($1, $2, $3, $4, $5, true, $6)
RETURNING id, client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at;
`
	var c model.VPNClient
	err := r.db.QueryRow(ctx, q,
		p.ClientUUID,
		p.ServerID,
		p.TelegramUserID,
		p.MaxIPs,
		p.KeyExpiresAt.UTC(),
		p.Note,
	).Scan(
		&c.ID,
		&c.ClientUUID,
		&c.ServerID,
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

func (r *VPNClientsRepo) ListMonitorTargets(ctx context.Context, now time.Time) ([]model.MonitorTarget, error) {
	const q = `
SELECT c.client_uuid, c.max_ips, xa.xui_client_email
FROM vpn_clients c
INNER JOIN xui_access xa ON xa.client_uuid = c.client_uuid
WHERE c.is_active = true AND c.key_expires_at > $1;
`
	rows, err := r.db.Query(ctx, q, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.MonitorTarget
	for rows.Next() {
		var t model.MonitorTarget
		if err := rows.Scan(&t.ClientUUID, &t.MaxIPs, &t.XUIClientEmail); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
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

func (r *VPNClientsRepo) ListActiveByNote(ctx context.Context, note string) ([]model.VPNClient, error) {
	const q = `
SELECT
  id, client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at
FROM vpn_clients
WHERE is_active = true AND note = $1
ORDER BY created_at DESC;
`
	rows, err := r.db.Query(ctx, q, strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.VPNClient
	for rows.Next() {
		var c model.VPNClient
		if err := rows.Scan(
			&c.ID,
			&c.ClientUUID,
			&c.ServerID,
			&c.TelegramUserID,
			&c.MaxIPs,
			&c.KeyExpiresAt,
			&c.IsActive,
			&c.Note,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *VPNClientsRepo) ExtendKeyExpiresAt(ctx context.Context, clientUUID uuid.UUID, addDays int, now time.Time) (model.VPNClient, error) {
	if addDays <= 0 {
		return model.VPNClient{}, usecase.ErrInvalidExtend
	}
	c, err := r.GetByUUID(ctx, clientUUID)
	if err != nil {
		return model.VPNClient{}, err
	}
	if !c.IsActive {
		return model.VPNClient{}, usecase.ErrInactive
	}
	base := c.KeyExpiresAt.UTC()
	now = now.UTC()
	if base.Before(now) {
		base = now
	}
	newExpires := base.Add(time.Duration(addDays) * 24 * time.Hour)

	const q = `
UPDATE vpn_clients
SET key_expires_at = $2, updated_at = now()
WHERE client_uuid = $1
RETURNING id, client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at;
`
	var out model.VPNClient
	err = r.db.QueryRow(ctx, q, clientUUID, newExpires).Scan(
		&out.ID,
		&out.ClientUUID,
		&out.ServerID,
		&out.TelegramUserID,
		&out.MaxIPs,
		&out.KeyExpiresAt,
		&out.IsActive,
		&out.Note,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *VPNClientsRepo) UpdateMaxIPs(ctx context.Context, clientUUID uuid.UUID, maxIPs int) (model.VPNClient, error) {
	if maxIPs <= 0 {
		return model.VPNClient{}, usecase.ErrInvalidMaxIPs
	}
	const q = `
UPDATE vpn_clients
SET max_ips = $2, updated_at = now()
WHERE client_uuid = $1 AND is_active = true
RETURNING id, client_uuid, server_id, telegram_user_id, max_ips, key_expires_at, is_active, note, created_at, updated_at;
`
	var c model.VPNClient
	err := r.db.QueryRow(ctx, q, clientUUID, maxIPs).Scan(
		&c.ID,
		&c.ClientUUID,
		&c.ServerID,
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
