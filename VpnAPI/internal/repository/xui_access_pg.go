package repository

import (
	"api-vpn/internal/model"
	"api-vpn/internal/usecase"
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type XUIAccessRepo struct {
	db *pgxpool.Pool
}

func NewXUIAccessRepo(db *pgxpool.Pool) *XUIAccessRepo {
	return &XUIAccessRepo{db: db}
}

func (r *XUIAccessRepo) GetByClientUUID(ctx context.Context, clientUUID uuid.UUID) (model.XUIAccess, error) {
	const q = `
SELECT id, client_uuid, inbound_id, xui_client_email, vless_uri, created_at, updated_at
FROM xui_access
WHERE client_uuid = $1
LIMIT 1;
`
	var a model.XUIAccess
	err := r.db.QueryRow(ctx, q, clientUUID).Scan(
		&a.ID,
		&a.ClientUUID,
		&a.InboundID,
		&a.XUIClientEmail,
		&a.VLESSURI,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.XUIAccess{}, usecase.ErrNotFound
	}
	return a, err
}

func (r *XUIAccessRepo) Upsert(ctx context.Context, p model.UpsertXUIAccessParams) (model.XUIAccess, error) {
	const q = `
INSERT INTO xui_access (client_uuid, inbound_id, xui_client_email, vless_uri)
VALUES ($1, $2, $3, $4)
ON CONFLICT (client_uuid)
DO UPDATE SET inbound_id = EXCLUDED.inbound_id,
              xui_client_email = EXCLUDED.xui_client_email,
              vless_uri = EXCLUDED.vless_uri,
              updated_at = now()
RETURNING id, client_uuid, inbound_id, xui_client_email, vless_uri, created_at, updated_at;
`
	var a model.XUIAccess
	err := r.db.QueryRow(ctx, q, p.ClientUUID, p.InboundID, p.XUIClientEmail, p.VLESSURI).Scan(
		&a.ID,
		&a.ClientUUID,
		&a.InboundID,
		&a.XUIClientEmail,
		&a.VLESSURI,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	return a, err
}

func (r *XUIAccessRepo) DeleteByClientUUID(ctx context.Context, clientUUID uuid.UUID) error {
	const q = `DELETE FROM xui_access WHERE client_uuid = $1;`
	ct, err := r.db.Exec(ctx, q, clientUUID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return usecase.ErrNotFound
	}
	return nil
}

