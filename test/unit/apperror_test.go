package unit_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
)

func TestStatusCodeReturnsAppErrorStatus(t *testing.T) {
	t.Parallel()

	err := apperror.BadRequest("invalid payload")

	if got := apperror.StatusCode(err); got != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, got)
	}
}

func TestStatusCodeReturnsInternalServerErrorForGenericError(t *testing.T) {
	t.Parallel()

	err := errors.New("boom")

	if got := apperror.StatusCode(err); got != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, got)
	}
}
