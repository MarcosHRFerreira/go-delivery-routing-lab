package integration_test

import (
	"net/http"
	"testing"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type createBuyerRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
}

func newValidationTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	validate := validator.New()
	router.POST("/buyers", func(c *gin.Context) {
		var req createBuyerRequest
		if !httpresponse.BindAndValidateJSON(c, validate, &req) {
			return
		}

		httpresponse.JSON(c, http.StatusCreated, gin.H{
			"message": "buyer created",
		})
	})

	return router
}

func TestBindAndValidateJSONReturnsBadRequestForMalformedJSON(t *testing.T) {
	t.Parallel()

	router := newValidationTestRouter()
	recorder := performRawRequest(t, router, http.MethodPost, "/buyers", "{invalid-json", map[string]string{
		"Content-Type": "application/json",
	})

	assertStatusCode(t, recorder, http.StatusBadRequest)
	assertJSONMessage(t, recorder, "invalid request body")
}

func TestBindAndValidateJSONReturnsValidationErrorsForInvalidPayload(t *testing.T) {
	t.Parallel()

	router := newValidationTestRouter()
	recorder := performJSONRequest(t, router, http.MethodPost, "/buyers", map[string]string{
		"name":  "Ma",
		"email": "invalid-email",
	}, nil)

	assertStatusCode(t, recorder, http.StatusBadRequest)
	assertJSONMessage(t, recorder, "validation failed")

	payload := decodeJSONResponse(t, recorder)
	errorsField, ok := payload["errors"].([]interface{})
	if !ok {
		t.Fatalf("expected validation errors array, got %T", payload["errors"])
	}

	if len(errorsField) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(errorsField))
	}
}
