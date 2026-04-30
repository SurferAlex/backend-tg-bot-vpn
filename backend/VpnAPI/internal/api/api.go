package api

import (
	"api-vpn/internal/handlers"
	"api-vpn/internal/middleware"
	"api-vpn/internal/repository"
	"api-vpn/internal/usecase"
	"api-vpn/internal/xui"
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type XUISetup struct {
	BaseURL            string
	Username           string
	Password           string
	InboundID          int64
	ExternalHost       string
	Fingerprint        string
	SpiderX            string
	Flow               string
	HostHeader         string
	ServerName         string
	InsecureSkipVerify bool
}

func SetupServer(db *pgxpool.Pool, internalToken string, xuiCfg XUISetup) *gin.Engine {
	r := gin.Default()

	repo := repository.NewVPNClientsRepo(db)
	uc := usecase.NewVPNClients(repo)

	xuiClient := xui.New(xuiCfg.BaseURL, xuiCfg.Username, xuiCfg.Password)
	if xuiCfg.HostHeader != "" {
		xuiClient.WithHostHeader(xuiCfg.HostHeader)
	}
	if strings.HasPrefix(strings.ToLower(xuiCfg.BaseURL), "https://") && (xuiCfg.InsecureSkipVerify || xuiCfg.ServerName != "") {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: xuiCfg.InsecureSkipVerify,
		}
		if xuiCfg.ServerName != "" {
			tlsCfg.ServerName = xuiCfg.ServerName
		}
		tr := &http.Transport{TLSClientConfig: tlsCfg}
		xuiClient.WithHTTPClient(&http.Client{Timeout: 10 * time.Second, Transport: tr})
	}
	xuiRepo := repository.NewXUIAccessRepo(db)
	xuiUC := usecase.NewXUIAccess(xuiRepo, uc, xuiClient, xuiCfg.InboundID, xuiCfg.ExternalHost, xuiCfg.Fingerprint, xuiCfg.SpiderX, xuiCfg.Flow)

	h := &handlers.Handlers{
		DB:        db,
		Clients:   uc,
		XUIAccess: xuiUC,
	}

	RegisterRoutes(r, h, internalToken)

	return r
}

func RegisterRoutes(r *gin.Engine, h *handlers.Handlers, internalToken string) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", h.Ping)
		v1.GET("/health", h.Health)
	}

	protected := v1.Group("")
	protected.Use(middleware.InternalToken(internalToken))
	{
		protected.POST("/clients", h.CreateClient)
		protected.GET("/clients/:uuid", h.GetClient)
		protected.POST("/clients/:uuid/deactivate", h.DeactivateClient)
		// protected.POST("/clients/:uuid/provision", h.ProvisionAccess)
		// protected.GET("/clients/:uuid/access", h.GetAccess)
		// protected.POST("/clients/:uuid/revoke", h.RevokeAccess)
	}
}
