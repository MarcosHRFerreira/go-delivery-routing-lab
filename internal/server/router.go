package server

import (
	"database/sql"
	"time"

	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	healthhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/health"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const healthCheckTimeout = 2 * time.Second

type Dependencies struct {
	DB *sql.DB
	HealthChecker healthhandler.Checker
}

func NewRouter(validate *validator.Validate, deps Dependencies) *gin.Engine {
	if validate == nil {
		validate = validator.New()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	healthhandler.NewHandler(router, deps.HealthChecker, healthCheckTimeout).RouteList()

	if deps.DB == nil{
		return router
	}

	db:= deps.DB
	buyerRepo := buyerrepository.NewRepository(db)
	buyerService := buyerservice.NewService(buyerRepo)
	buyerHandler := buyerhandler.NewHandler(router, validate, buyerService)
	buyerHandler.RouteList()
	
	return router
}
