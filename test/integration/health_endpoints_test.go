package integration_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestHealthEndpointReturnsOKWhenDatabasePingSucceeds(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&healthCheckerStub{
		pingContextFunc: func(ctx context.Context) error {
			return nil
		},
	})

	recorder := performJSONRequest(t, router, http.MethodGet, "/check-health", nil, nil)

	assertStatusCode(t, recorder, http.StatusOK)
	assertJSONMessage(t, recorder, "service is healthy")
}

func TestHealthEndpointReturnsServiceUnavailableWhenDatabasePingFails(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&healthCheckerStub{
		pingContextFunc: func(ctx context.Context) error {
			return errors.New("db down")
		},
	})

	recorder := performJSONRequest(t, router, http.MethodGet, "/check-health", nil, nil)

	assertStatusCode(t, recorder, http.StatusServiceUnavailable)
	assertJSONMessage(t, recorder, "database unavailable")
}
