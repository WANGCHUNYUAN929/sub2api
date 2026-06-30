package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/http/httpguts"
)

const AccountRequestHeadersOverrideExtraKey = "request_headers_override"

var deniedAccountRequestHeaderOverride = map[string]struct{}{
	"authorization":       {},
	"cookie":              {},
	"host":                {},
	"content-length":      {},
	"transfer-encoding":   {},
	"connection":          {},
	"proxy-authorization": {},
	"x-api-key":           {},
	"x-goog-api-key":      {},
}

var allowedAccountRequestHeaderOverride = map[string]struct{}{
	"user-agent": {},
}

func ValidateAccountRequestHeadersOverride(extra map[string]any) error {
	if extra == nil {
		return nil
	}
	headers, ok, err := accountRequestHeadersOverrideFromExtra(extra)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	for key, value := range headers {
		if err := validateAccountRequestHeaderOverridePair(key, value); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAccountRequestHeadersOverrideForAccount(platform, accountType string, extra map[string]any) error {
	if extra == nil {
		return nil
	}
	_, ok, err := accountRequestHeadersOverrideFromExtra(extra)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if platform != PlatformOpenAI && !(platform == PlatformAnthropic && accountType == AccountTypeAPIKey) {
		return errors.New("request header override only supports OpenAI accounts or Anthropic API Key accounts")
	}
	return ValidateAccountRequestHeadersOverride(extra)
}

func ApplyAccountRequestHeadersOverride(req *http.Request, account *Account) {
	if req == nil || account == nil || account.Extra == nil {
		return
	}
	headers, ok, err := accountRequestHeadersOverrideFromExtra(account.Extra)
	if !ok || err != nil {
		return
	}
	for key, value := range headers {
		if validateAccountRequestHeaderOverridePair(key, value) != nil {
			continue
		}
		req.Header.Set(http.CanonicalHeaderKey(strings.TrimSpace(key)), strings.TrimSpace(value))
	}
}

func accountRequestHeadersOverrideFromExtra(extra map[string]any) (map[string]string, bool, error) {
	raw, ok := extra[AccountRequestHeadersOverrideExtraKey]
	if !ok || raw == nil {
		return nil, false, nil
	}
	switch typed := raw.(type) {
	case map[string]any:
		out := make(map[string]string, len(typed))
		for key, value := range typed {
			stringValue, ok := value.(string)
			if !ok {
				return nil, true, fmt.Errorf("%s header value must be string", AccountRequestHeadersOverrideExtraKey)
			}
			out[key] = stringValue
		}
		return out, true, nil
	case map[string]string:
		out := make(map[string]string, len(typed))
		for key, value := range typed {
			out[key] = value
		}
		return out, true, nil
	default:
		return nil, true, fmt.Errorf("%s must be an object", AccountRequestHeadersOverrideExtraKey)
	}
}

func validateAccountRequestHeaderOverridePair(key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return errors.New("request header override name cannot be empty")
	}
	if value == "" {
		return fmt.Errorf("%s value cannot be empty", key)
	}
	if !httpguts.ValidHeaderFieldName(key) {
		return fmt.Errorf("invalid request header override name: %s", key)
	}
	if !httpguts.ValidHeaderFieldValue(value) {
		return fmt.Errorf("invalid request header override value for %s", key)
	}
	lower := strings.ToLower(key)
	if _, denied := deniedAccountRequestHeaderOverride[lower]; denied {
		return fmt.Errorf("request header override cannot set %s", key)
	}
	if _, allowed := allowedAccountRequestHeaderOverride[lower]; !allowed {
		return errors.New("request header override only supports User-Agent")
	}
	return nil
}
