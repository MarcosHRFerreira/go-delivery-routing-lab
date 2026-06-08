package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type healthCheckerStub struct {
	pingContextFunc func(ctx context.Context) error
}

func (s *healthCheckerStub) PingContext(ctx context.Context) error {
	if s.pingContextFunc == nil {
		return nil
	}

	return s.pingContextFunc(ctx)
}

func newTestRouter(checker *healthCheckerStub) *gin.Engine {
	gin.SetMode(gin.TestMode)

	return server.NewRouter(validator.New(), server.Dependencies{
		HealthChecker: checker,
	})
}

func performJSONRequest(t *testing.T, router *gin.Engine, method string, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func performRawRequest(t *testing.T, router *gin.Engine, method string, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func decodeJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var payload map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return payload
}

func assertStatusCode(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if recorder.Code != expected {
		t.Fatalf("expected status %d, got %d. body=%s", expected, recorder.Code, recorder.Body.String())
	}
}

func assertJSONMessage(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	t.Helper()

	payload := decodeJSONResponse(t, recorder)
	message, ok := payload["message"].(string)
	if !ok {
		t.Fatalf("expected string message in response, got %v", payload["message"])
	}

	if message != expected {
		t.Fatalf("expected message %q, got %q", expected, message)
	}
}
