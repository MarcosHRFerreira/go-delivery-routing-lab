package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetBuyerByID(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}
	response, err := h.buyerService.GetByID(c.Request.Context(), buyerID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	httpresponse.JSON(c, http.StatusOK, response)

}
