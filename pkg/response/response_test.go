package response

import (
	"encoding/json"
	"errors"
	"member-shop-api/pkg/apperror"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Trace-Id", "test-trace-123")

	Success(c, map[string]string{"foo": "bar"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
	if res.Code != apperror.CodeSuccess {
		t.Errorf("expected code %s, got %s", apperror.CodeSuccess, res.Code)
	}
	if res.TraceID != "test-trace-123" {
		t.Errorf("expected traceId test-trace-123, got %s", res.TraceID)
	}
}

func TestFailResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)

	appErr := apperror.New(http.StatusUnauthorized, apperror.CodeAuthInvalidCredentials, "Tên đăng nhập không đúng")
	Fail(c, appErr)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if res.Success {
		t.Errorf("expected success false, got true")
	}
	if res.Code != apperror.CodeAuthInvalidCredentials {
		t.Errorf("expected code %s, got %s", apperror.CodeAuthInvalidCredentials, res.Code)
	}
}

func TestHandleErrorFallback(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	rawErr := errors.New("database connection dropped")
	HandleError(c, rawErr)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should mask internal raw error message to client
	if res.Code != apperror.CodeInternalServerError {
		t.Errorf("expected code %s, got %s", apperror.CodeInternalServerError, res.Code)
	}
	if res.Message == "database connection dropped" {
		t.Errorf("expected masked message, but raw error leaked: %s", res.Message)
	}
}
