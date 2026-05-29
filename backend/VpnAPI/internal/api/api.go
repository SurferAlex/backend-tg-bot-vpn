package api

import (
	"api-vpn/internal/handlers"
	"api-vpn/internal/middleware"
	"api-vpn/internal/repository"
	"api-vpn/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupServer(db *pgxpool.Pool, internalToken string) *gin.Engine {
	r := gin.Default()

	serversRepo := repository.NewVPNServersRepo(db)
	serversUC := usecase.NewVPNServers(serversRepo)

	clientsRepo := repository.NewVPNClientsRepo(db)
	clientsUC := usecase.NewVPNClients(clientsRepo)

	registry := usecase.NewXUIRegistry(serversUC)
	xuiRepo := repository.NewXUIAccessRepo(db)
	xuiUC := usecase.NewXUIAccess(xuiRepo, clientsUC, registry)

	h := &handlers.Handlers{
		DB:        db,
		Servers:   serversUC,
		Clients:   clientsUC,
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
		protected.GET("/servers", h.ListServers)
		protected.POST("/clients", h.CreateClient)
		protected.GET("/clients/:uuid", h.GetClient)
		protected.POST("/clients/:uuid/deactivate", h.DeactivateClient)
		protected.POST("/clients/:uuid/provision", h.ProvisionAccess)
		protected.GET("/clients/:uuid/access", h.GetAccess)
		protected.POST("/clients/:uuid/revoke", h.RevokeAccess)
		protected.GET("/monitor/targets", h.ListMonitorTargets)
	}
}
