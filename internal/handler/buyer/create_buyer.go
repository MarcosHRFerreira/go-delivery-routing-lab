package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBuyer(c *gin.Context) {
	var req dto.CreateBuyerRequest

	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}
	buyerID, err := h.buyerService.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": buyerID,
	})

}
