package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllBuyers(c *gin.Context) {
	response, err := h.buyerService.GetAll(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}
	httpresponse.JSON(c, http.StatusOK, response)
}
