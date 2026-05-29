package model

import "time"

type VPNServer struct {
	ID                   string
	Name                 string
	IsActive             bool
	XUIBaseURL           string
	XUIUsername          string
	XUIPassword          string
	XUIInboundID         int64
	XUIExternalHost      string
	XUIFingerprint       string
	XUISpiderX           string
	XUIFlow              string
	XUIHostHeader        string
	XUIServerName        string
	XUIInsecureSkipVerify bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UpsertVPNServerParams struct {
	ID                   string
	Name                 string
	IsActive             bool
	XUIBaseURL           string
	XUIUsername          string
	XUIPassword          string
	XUIInboundID         int64
	XUIExternalHost      string
	XUIFingerprint       string
	XUISpiderX           string
	XUIFlow              string
	XUIHostHeader        string
	XUIServerName        string
	XUIInsecureSkipVerify bool
}
