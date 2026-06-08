package buyer

import (
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router       gin.IRouter
	validate     *validator.Validate
	buyerService buyerservice.BuyerService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, buyerService buyerservice.BuyerService) *Handler {
	return &Handler{
		router:   router,
		validate: validate,
		 buyerService: buyerService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/buyers", h.CreateBuyer)
	h.router.GET("/buyers", h.GetAllBuyers)
	h.router.GET("/buyers/:buyer_id", h.GetBuyerByID)
	h.router.PUT("/buyers/:buyer_id", h.UpdateBuyer)
	h.router.DELETE("/buyers/:buyer_id", h.DeleteBuyer)
}
