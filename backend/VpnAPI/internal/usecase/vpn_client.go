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
var ErrInvalidExtend = errors.New("invalid extend")
var ErrInvalidMaxIPs = errors.New("invalid max ips")
var ErrAmbiguousRef = errors.New("ambiguous client ref")

type VPNClientsRepo interface {
	GetByUUID(ctx context.Context, id uuid.UUID) (model.VPNClient, error)
	Create(ctx context.Context, p model.CreateVPNClientParams) (model.VPNClient, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	ExtendKeyExpiresAt(ctx context.Context, id uuid.UUID, addDays int, now time.Time) (model.VPNClient, error)
	UpdateMaxIPs(ctx context.Context, id uuid.UUID, maxIPs int) (model.VPNClient, error)
	ListActiveByNote(ctx context.Context, note string) ([]model.VPNClient, error)
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

func (uc *VPNClients) Extend(ctx context.Context, id uuid.UUID, addDays int) (model.VPNClient, error) {
	if addDays <= 0 {
		return model.VPNClient{}, ErrInvalidExtend
	}
	return uc.repo.ExtendKeyExpiresAt(ctx, id, addDays, uc.now())
}

func (uc *VPNClients) UpdateMaxIPs(ctx context.Context, id uuid.UUID, maxIPs int) (model.VPNClient, error) {
	if maxIPs < 1 || maxIPs > 6 {
		return model.VPNClient{}, ErrInvalidMaxIPs
	}
	return uc.repo.UpdateMaxIPs(ctx, id, maxIPs)
}

func (uc *VPNClients) GetByUUID(ctx context.Context, id uuid.UUID) (model.VPNClient, error) {
	return uc.repo.GetByUUID(ctx, id)
}

// ResolveRef accepts a client UUID or exact client name (note field).
func (uc *VPNClients) ResolveRef(ctx context.Context, ref string) (model.VPNClient, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return model.VPNClient{}, ErrNotFound
	}
	if id, err := uuid.FromString(ref); err == nil {
		return uc.repo.GetByUUID(ctx, id)
	}
	list, err := uc.repo.ListActiveByNote(ctx, ref)
	if err != nil {
		return model.VPNClient{}, err
	}
	switch len(list) {
	case 0:
		return model.VPNClient{}, ErrNotFound
	case 1:
		return list[0], nil
	default:
		return model.VPNClient{}, ErrAmbiguousRef
	}
}

func (uc *VPNClients) ListMonitorTargets(ctx context.Context) ([]model.MonitorTarget, error) {
	return uc.repo.ListMonitorTargets(ctx, uc.now())
}
