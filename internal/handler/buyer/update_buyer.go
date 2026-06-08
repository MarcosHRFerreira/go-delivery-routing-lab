package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateBuyer(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}
	var req dto.UpdateBuyerRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}
	response, err := h.buyerService.Update(c.Request.Context(), buyerID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	httpresponse.JSON(c, http.StatusOK, response)

}
