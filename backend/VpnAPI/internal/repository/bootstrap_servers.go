package repository

import (
	"context"
	"log"

	"api-vpn/internal/config"
	"api-vpn/internal/model"
)

// SyncDefaultServerFromEnv upserts server "default" from legacy XUI_* env (control-plane primary panel).
func SyncDefaultServerFromEnv(ctx context.Context, repo *VPNServersRepo, cfg config.XUIConfig) error {
	return repo.Upsert(ctx, model.UpsertVPNServerParams{
		ID:                    "default",
		Name:                  "Default",
		IsActive:              true,
		XUIBaseURL:            cfg.BaseURL,
		XUIUsername:           cfg.Username,
		XUIPassword:           cfg.Password,
		XUIInboundID:          cfg.InboundID,
		XUIExternalHost:       cfg.ExternalHost,
		XUIFingerprint:        cfg.Fingerprint,
		XUISpiderX:            cfg.SpiderX,
		XUIFlow:               cfg.Flow,
		XUIHostHeader:         cfg.HostHeader,
		XUIServerName:         cfg.ServerName,
		XUIInsecureSkipVerify: cfg.InsecureSkipVerify,
	})
}

func optionalServerParams(s config.OptionalVPNServer) model.UpsertVPNServerParams {
	return model.UpsertVPNServerParams{
		ID:                    s.ID,
		Name:                  s.Name,
		IsActive:              true,
		XUIBaseURL:            s.XUI.BaseURL,
		XUIUsername:           s.XUI.Username,
		XUIPassword:           s.XUI.Password,
		XUIInboundID:          s.XUI.InboundID,
		XUIExternalHost:       s.XUI.ExternalHost,
		XUIFingerprint:        s.XUI.Fingerprint,
		XUISpiderX:            s.XUI.SpiderX,
		XUIFlow:               s.XUI.Flow,
		XUIHostHeader:         s.XUI.HostHeader,
		XUIServerName:         s.XUI.ServerName,
		XUIInsecureSkipVerify: s.XUI.InsecureSkipVerify,
	}
}

// EnsureOptionalServersFromEnv inserts extra servers from env when their id is not in DB yet.
func EnsureOptionalServersFromEnv(ctx context.Context, repo *VPNServersRepo, servers []config.OptionalVPNServer) error {
	for _, s := range servers {
		created, err := repo.InsertIfNotExists(ctx, optionalServerParams(s))
		if err != nil {
			return err
		}
		if created {
			log.Printf("vpn server bootstrapped from env: id=%s name=%s", s.ID, s.Name)
		}
	}
	return nil
}
