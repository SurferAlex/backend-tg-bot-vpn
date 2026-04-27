package usecase

import (
	"api-vpn/internal/model"
	"api-vpn/internal/xui"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

type XUIAccessRepo interface {
	GetByClientUUID(ctx context.Context, clientUUID uuid.UUID) (model.XUIAccess, error)
	Upsert(ctx context.Context, p model.UpsertXUIAccessParams) (model.XUIAccess, error)
	DeleteByClientUUID(ctx context.Context, clientUUID uuid.UUID) error
}

type XUIAccess struct {
	repo      XUIAccessRepo
	clientsUC *VPNClients
	xui       *xui.Client
	inboundID int64
	external  string
	fp        string
	spiderX   string
	flow      string
	now       func() time.Time
}

func NewXUIAccess(repo XUIAccessRepo, clientsUC *VPNClients, xuiClient *xui.Client, inboundID int64, externalHost, fp, spiderX, flow string) *XUIAccess {
	return &XUIAccess{
		repo:      repo,
		clientsUC: clientsUC,
		xui:       xuiClient,
		inboundID: inboundID,
		external:  externalHost,
		fp:        fp,
		spiderX:   spiderX,
		flow:      flow,
		now:       time.Now,
	}
}

func (uc *XUIAccess) Get(ctx context.Context, clientUUID uuid.UUID) (model.XUIAccess, error) {
	return uc.repo.GetByClientUUID(ctx, clientUUID)
}

func (uc *XUIAccess) Provision(ctx context.Context, clientUUID uuid.UUID) (model.XUIAccess, error) {
	client, err := uc.clientsUC.GetActiveByUUID(ctx, clientUUID)
	if err != nil {
		return model.XUIAccess{}, err
	}

	displayName := client.ClientUUID.String()
	if client.Note != nil {
		if n := strings.TrimSpace(*client.Note); n != "" {
			displayName = n
		}
	}
	xuiEmail := makeXUIEmail(displayName, client.ClientUUID.String())
	expiryMs := client.KeyExpiresAt.UTC().UnixMilli()
	limitIP := client.MaxIPs

	if err := uc.xui.AddOrUpdateVLESSClient(ctx, uc.inboundID, client.ClientUUID.String(), xuiEmail, limitIP, expiryMs, uc.flow); err != nil {
		return model.XUIAccess{}, err
	}

	inb, ss, err := uc.xui.GetInbound(ctx, uc.inboundID)
	if err != nil {
		return model.XUIAccess{}, err
	}
	uri, err := xui.BuildVLESSRealityURI(uc.external, inb.Port, client.ClientUUID.String(), displayName, ss, uc.fp, uc.spiderX, uc.flow)
	if err != nil {
		return model.XUIAccess{}, err
	}

	return uc.repo.Upsert(ctx, model.UpsertXUIAccessParams{
		ClientUUID:     client.ClientUUID,
		InboundID:      uc.inboundID,
		XUIClientEmail: xuiEmail,
		VLESSURI:       uri,
	})
}

var xuiEmailAllowed = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func makeXUIEmail(displayName string, uuidStr string) string {
	base := strings.TrimSpace(displayName)
	if base == "" {
		base = uuidStr
	}
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, " ", "_")
	base = xuiEmailAllowed.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._-")

	suffix := uuidStr
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	out := base
	if base != uuidStr {
		out = base + "-" + suffix
	}
	if out == "" {
		out = uuidStr
	}

	if len(out) > 64 {
		out = out[:64]
	}
	return out
}

func (uc *XUIAccess) Revoke(ctx context.Context, clientUUID uuid.UUID) error {
	a, err := uc.repo.GetByClientUUID(ctx, clientUUID)
	if err != nil {
		return err
	}
	if err := uc.xui.DeleteClientByEmail(ctx, a.InboundID, a.XUIClientEmail); err != nil {
		return fmt.Errorf("xui delete: %w", err)
	}
	_ = uc.repo.DeleteByClientUUID(ctx, clientUUID)
	return nil
}
