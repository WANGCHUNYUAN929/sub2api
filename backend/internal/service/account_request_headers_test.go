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

func TestWithAccountTestRequestHeadersOverrideDoesNotMutateOriginal(t *testing.T) {
	account := &Account{
		Extra: map[string]any{
			"kept": "value",
		},
	}

	got := withAccountTestRequestHeadersOverride(account, map[string]string{
		"User-Agent": "claude-cli/2.1.196 (external, claude-vscode, agent-sdk/0.3.196)",
	})

	require.NotSame(t, account, got)
	require.Equal(t, "value", got.Extra["kept"])
	require.Nil(t, account.Extra[AccountRequestHeadersOverrideExtraKey])
	require.Equal(t, map[string]any{
		"User-Agent": "claude-cli/2.1.196 (external, claude-vscode, agent-sdk/0.3.196)",
	}, got.Extra[AccountRequestHeadersOverrideExtraKey])
}
