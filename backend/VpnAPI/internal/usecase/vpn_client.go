package usecase

import (
	"api-vpn/internal/model"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

var ErrNotFound = errors.New("not found")
var ErrInactive = errors.New("inactive")
var ErrExpired = errors.New("expired")
var ErrInvalidServer = errors.New("invalid server")

type VPNClientsRepo interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (model.VPNClient, error)
	Create(ctx context.Context, p model.CreateVPNClientParams) (model.VPNClient, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	ListMonitorTargets(ctx context.Context, now time.Time) ([]model.MonitorTarget, error)
}
type VPNClients struct {
	repo VPNClientsRepo
	now  func() time.Time
}

func NewVPNClients(repo VPNClientsRepo) *VPNClients {
	return &VPNClients{repo: repo, now: time.Now}
}

func (uc *VPNClients) GetActiveByUUID(ctx context.Context, id uuid.UUID) (model.VPNClient, error) {
	c, err := uc.repo.GetByUUID(ctx, id)
	if err != nil {
		return model.VPNClient{}, err
	}
	if !c.IsActive {
		return model.VPNClient{}, ErrInactive
	}
	if !c.KeyExpiresAt.After(uc.now()) {
		return model.VPNClient{}, ErrExpired
	}
	return c, nil
}

func (uc *VPNClients) Create(ctx context.Context, p model.CreateVPNClientParams) (model.VPNClient, error) {
	if p.MaxIPs <= 0 {
		p.MaxIPs = 2
	}
	p.ServerID = strings.TrimSpace(p.ServerID)
	if p.ServerID == "" {
		return model.VPNClient{}, ErrInvalidServer
	}
	return uc.repo.Create(ctx, p)
}
func (uc *VPNClients) Deactivate(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Deactivate(ctx, id)
}
func (uc *VPNClients) GetByUUID(ctx context.Context, id uuid.UUID) (model.VPNClient, error) {
	return uc.repo.GetByUUID(ctx, id)
}

func (uc *VPNClients) ListMonitorTargets(ctx context.Context) ([]model.MonitorTarget, error) {
	return uc.repo.ListMonitorTargets(ctx, uc.now())
}
