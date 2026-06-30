//go:build unit

package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerCreateRequestHeadersOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	handler := NewAccountHandler(
		adminSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	router := gin.New()
	router.POST("/api/v1/admin/accounts", handler.Create)

	body := map[string]any{
		"name":     "openai-1",
		"platform": "openai",
		"type":     "apikey",
		"credentials": map[string]any{
			"api_key": "sk-test",
		},
		"extra": map[string]any{
			"request_headers_override": map[string]any{
				"User-Agent": "codex_vscode/0.142.3",
			},
		},
		"concurrency": 1,
		"priority":    1,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.createdAccounts, 1)
	require.Equal(t, map[string]any{"User-Agent": "codex_vscode/0.142.3"}, adminSvc.createdAccounts[0].Extra["request_headers_override"])
}

func TestAccountHandlerCreateRejectsDeniedRequestHeadersOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	handler := NewAccountHandler(
		adminSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	router := gin.New()
	router.POST("/api/v1/admin/accounts", handler.Create)

	body := map[string]any{
		"name":     "openai-1",
		"platform": "openai",
		"type":     "apikey",
		"credentials": map[string]any{
			"api_key": "sk-test",
		},
		"extra": map[string]any{
			"request_headers_override": map[string]any{
				"Authorization": "Bearer bad",
			},
		},
		"concurrency": 1,
		"priority":    1,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, adminSvc.createdAccounts)
}

func TestAccountHandlerCreateRejectsUnsupportedPlatformRequestHeadersOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	handler := NewAccountHandler(
		adminSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	router := gin.New()
	router.POST("/api/v1/admin/accounts", handler.Create)

	body := map[string]any{
		"name":     "gemini-1",
		"platform": "gemini",
		"type":     "apikey",
		"credentials": map[string]any{
			"api_key": "test",
		},
		"extra": map[string]any{
			"request_headers_override": map[string]any{
				"User-Agent": "codex_vscode/0.142.3",
			},
		},
		"concurrency": 1,
		"priority":    1,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, adminSvc.createdAccounts)
}

func TestAccountHandlerUpdateRequestHeadersOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       7,
			Name:     "openai-1",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
		},
	}
	handler := NewAccountHandler(
		adminSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	router := gin.New()
	router.PUT("/api/v1/admin/accounts/:id", handler.Update)

	body := map[string]any{
		"name":   "openai-1",
		"status": "active",
		"extra": map[string]any{
			"request_headers_override": map[string]any{
				"User-Agent": "codex_vscode/0.142.3",
			},
		},
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/accounts/7", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.updatedAccounts, 1)
	require.Equal(t, map[string]any{"User-Agent": "codex_vscode/0.142.3"}, adminSvc.updatedAccounts[0].Extra["request_headers_override"])
}
