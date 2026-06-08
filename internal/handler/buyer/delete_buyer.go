package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteBuyer(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}
	if err := h.buyerService.Delete(c.Request.Context(), buyerID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	c.Status(http.StatusNoContent)

}
