//go:build unit

package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAccountRequestHeadersOverride(t *testing.T) {
	t.Run("allows user agent", func(t *testing.T) {
		err := ValidateAccountRequestHeadersOverride(map[string]any{
			AccountRequestHeadersOverrideExtraKey: map[string]any{
				"User-Agent": "codex_vscode/0.142.3",
			},
		})
		require.NoError(t, err)
	})

	t.Run("rejects auth header", func(t *testing.T) {
		err := ValidateAccountRequestHeadersOverride(map[string]any{
			AccountRequestHeadersOverrideExtraKey: map[string]any{
				"Authorization": "Bearer bad",
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "Authorization")
	})

	t.Run("rejects non string value", func(t *testing.T) {
		err := ValidateAccountRequestHeadersOverride(map[string]any{
			AccountRequestHeadersOverrideExtraKey: map[string]any{
				"User-Agent": 123,
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be string")
	})
}

func TestApplyAccountRequestHeadersOverride(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "original")

	account := &Account{
		Extra: map[string]any{
			AccountRequestHeadersOverrideExtraKey: map[string]any{
				"User-Agent": "codex_vscode/0.142.3",
			},
		},
	}

	ApplyAccountRequestHeadersOverride(req, account)

	require.Equal(t, "codex_vscode/0.142.3", req.Header.Get("User-Agent"))
}
