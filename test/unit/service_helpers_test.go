package unit_test

import (
	"testing"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
)

func assertStatusCode(t *testing.T, err error, expected int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error with status %d, got nil", expected)
	}

	if actual := apperror.StatusCode(err); actual != expected {
		t.Fatalf("expected status %d, got %d", expected, actual)
	}
}

func assertErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error with message %q, got nil", expected)
	}

	if err.Error() != expected {
		t.Fatalf("expected message %q, got %q", expected, err.Error())
	}
}
