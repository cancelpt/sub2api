//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardAsAnthropic_StrictRetryMarksUnhandledCustomErrorsRetryableOnSameAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"model":"gpt-5.4","max_tokens":16,"messages":[{"role":"user","content":"hello"}],"stream":false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"rid_strict_retry"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"server exploded"}}`)),
	}}
	settingSvc := NewSettingService(&strictSchedulerSettingsRepoStub{
		all: map[string]string{
			SettingKeyOpenAIStrictSchedulerEnabled: "true",
			"openai_strict_retry_enabled":          "true",
			"openai_strict_retry_count":            "4",
		},
	}, &config.Config{})
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		httpUpstream:   upstream,
		settingService: settingSvc,
	}
	account := &Account{
		ID:          1,
		Name:        "xcode",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":                    "sk-test",
			"custom_error_codes_enabled": true,
			"custom_error_codes":         []any{float64(http.StatusForbidden)},
		},
	}

	_, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.True(t, failoverErr.RetryableOnSameAccount)
}

func TestForwardAsAnthropic_StrictRetrySkipsHandledCustomErrorCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"model":"gpt-5.4","max_tokens":16,"messages":[{"role":"user","content":"hello"}],"stream":false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"rid_strict_retry_403"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"forbidden"}}`)),
	}}
	settingSvc := NewSettingService(&strictSchedulerSettingsRepoStub{
		all: map[string]string{
			SettingKeyOpenAIStrictSchedulerEnabled: "true",
			"openai_strict_retry_enabled":          "true",
			"openai_strict_retry_count":            "4",
		},
	}, &config.Config{})
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		httpUpstream:   upstream,
		settingService: settingSvc,
	}
	account := &Account{
		ID:          1,
		Name:        "xcode",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":                    "sk-test",
			"custom_error_codes_enabled": true,
			"custom_error_codes":         []any{float64(http.StatusForbidden)},
		},
	}

	_, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.False(t, failoverErr.RetryableOnSameAccount)
}
