package usecase

import (
	"api-vpn/internal/model"
	"context"
	"strings"
)

type VPNServersRepo interface {
	GetByID(ctx context.Context, id string) (model.VPNServer, error)
	ListActive(ctx context.Context) ([]model.VPNServer, error)
	Upsert(ctx context.Context, p model.UpsertVPNServerParams) error
}

type VPNServers struct {
	repo VPNServersRepo
}

func NewVPNServers(repo VPNServersRepo) *VPNServers {
	return &VPNServers{repo: repo}
}

func (uc *VPNServers) ListActive(ctx context.Context) ([]model.VPNServer, error) {
	return uc.repo.ListActive(ctx)
}

func (uc *VPNServers) GetActiveByID(ctx context.Context, id string) (model.VPNServer, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.VPNServer{}, ErrInvalidServer
	}
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return model.VPNServer{}, err
	}
	if !s.IsActive {
		return model.VPNServer{}, ErrInvalidServer
	}
	return s, nil
}

func (uc *VPNServers) Upsert(ctx context.Context, p model.UpsertVPNServerParams) error {
	return uc.repo.Upsert(ctx, p)
}
