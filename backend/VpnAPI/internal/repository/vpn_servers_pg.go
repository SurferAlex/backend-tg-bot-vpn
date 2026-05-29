package repository

import (
	"context"
	"errors"

	"api-vpn/internal/model"
	"api-vpn/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VPNServersRepo struct {
	db *pgxpool.Pool
}

func NewVPNServersRepo(db *pgxpool.Pool) *VPNServersRepo {
	return &VPNServersRepo{db: db}
}

const vpnServerSelectCols = `
  id, name, is_active,
  xui_base_url, xui_username, xui_password, xui_inbound_id, xui_external_host,
  xui_fingerprint, xui_spiderx, xui_flow, xui_host_header, xui_server_name, xui_insecure_skip_verify,
  created_at, updated_at
`

func scanVPNServer(row pgx.Row) (model.VPNServer, error) {
	var s model.VPNServer
	err := row.Scan(
		&s.ID,
		&s.Name,
		&s.IsActive,
		&s.XUIBaseURL,
		&s.XUIUsername,
		&s.XUIPassword,
		&s.XUIInboundID,
		&s.XUIExternalHost,
		&s.XUIFingerprint,
		&s.XUISpiderX,
		&s.XUIFlow,
		&s.XUIHostHeader,
		&s.XUIServerName,
		&s.XUIInsecureSkipVerify,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	return s, err
}

func (r *VPNServersRepo) GetByID(ctx context.Context, id string) (model.VPNServer, error) {
	const q = `SELECT` + vpnServerSelectCols + ` FROM vpn_servers WHERE id = $1 LIMIT 1;`
	s, err := scanVPNServer(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.VPNServer{}, usecase.ErrNotFound
	}
	return s, err
}

func (r *VPNServersRepo) ListActive(ctx context.Context) ([]model.VPNServer, error) {
	const q = `SELECT` + vpnServerSelectCols + ` FROM vpn_servers WHERE is_active = true ORDER BY name;`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.VPNServer
	for rows.Next() {
		s, err := scanVPNServer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *VPNServersRepo) Upsert(ctx context.Context, p model.UpsertVPNServerParams) error {
	const q = `
INSERT INTO vpn_servers (
  id, name, is_active,
  xui_base_url, xui_username, xui_password, xui_inbound_id, xui_external_host,
  xui_fingerprint, xui_spiderx, xui_flow, xui_host_header, xui_server_name, xui_insecure_skip_verify
) VALUES (
  $1, $2, $3,
  $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13, $14
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  is_active = EXCLUDED.is_active,
  xui_base_url = EXCLUDED.xui_base_url,
  xui_username = EXCLUDED.xui_username,
  xui_password = EXCLUDED.xui_password,
  xui_inbound_id = EXCLUDED.xui_inbound_id,
  xui_external_host = EXCLUDED.xui_external_host,
  xui_fingerprint = EXCLUDED.xui_fingerprint,
  xui_spiderx = EXCLUDED.xui_spiderx,
  xui_flow = EXCLUDED.xui_flow,
  xui_host_header = EXCLUDED.xui_host_header,
  xui_server_name = EXCLUDED.xui_server_name,
  xui_insecure_skip_verify = EXCLUDED.xui_insecure_skip_verify,
  updated_at = now();
`
	_, err := r.db.Exec(ctx, q,
		p.ID,
		p.Name,
		p.IsActive,
		p.XUIBaseURL,
		p.XUIUsername,
		p.XUIPassword,
		p.XUIInboundID,
		p.XUIExternalHost,
		p.XUIFingerprint,
		p.XUISpiderX,
		p.XUIFlow,
		p.XUIHostHeader,
		p.XUIServerName,
		p.XUIInsecureSkipVerify,
	)
	return err
}

// InsertIfNotExists creates a server row only when id is not present yet.
func (r *VPNServersRepo) InsertIfNotExists(ctx context.Context, p model.UpsertVPNServerParams) (bool, error) {
	const q = `
INSERT INTO vpn_servers (
  id, name, is_active,
  xui_base_url, xui_username, xui_password, xui_inbound_id, xui_external_host,
  xui_fingerprint, xui_spiderx, xui_flow, xui_host_header, xui_server_name, xui_insecure_skip_verify
) VALUES (
  $1, $2, $3,
  $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13, $14
)
ON CONFLICT (id) DO NOTHING;
`
	ct, err := r.db.Exec(ctx, q,
		p.ID,
		p.Name,
		p.IsActive,
		p.XUIBaseURL,
		p.XUIUsername,
		p.XUIPassword,
		p.XUIInboundID,
		p.XUIExternalHost,
		p.XUIFingerprint,
		p.XUISpiderX,
		p.XUIFlow,
		p.XUIHostHeader,
		p.XUIServerName,
		p.XUIInsecureSkipVerify,
	)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
