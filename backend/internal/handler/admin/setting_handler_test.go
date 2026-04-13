//go:build unit

package admin

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type settingHandlerRepoStub struct {
	all map[string]string
}

func (s *settingHandlerRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	value, ok := s.all[key]
	if !ok {
		return nil, service.ErrSettingNotFound
	}
	return &service.Setting{Key: key, Value: value}, nil
}

func (s *settingHandlerRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	value, ok := s.all[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (s *settingHandlerRepoStub) Set(ctx context.Context, key, value string) error {
	if s.all == nil {
		s.all = map[string]string{}
	}
	s.all[key] = value
	return nil
}

func (s *settingHandlerRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = s.all[key]
	}
	return out, nil
}

func (s *settingHandlerRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	if s.all == nil {
		s.all = map[string]string{}
	}
	for key, value := range settings {
		s.all[key] = value
	}
	return nil
}

func (s *settingHandlerRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.all))
	for key, value := range s.all {
		out[key] = value
	}
	return out, nil
}

func (s *settingHandlerRepoStub) Delete(ctx context.Context, key string) error {
	delete(s.all, key)
	return nil
}

func TestSettingHandler_UpdateSettings_PreservesOpenAIStrictSchedulerWhenFieldOmitted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &settingHandlerRepoStub{
		all: map[string]string{
			service.SettingKeyRegistrationEnabled:          "true",
			service.SettingKeyOpenAIStrictSchedulerEnabled: "true",
		},
	}
	settingService := service.NewSettingService(repo, &config.Config{})
	handler := NewSettingHandler(settingService, nil, nil, nil, nil, nil)

	router := gin.New()
	router.PUT("/settings", handler.UpdateSettings)

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, "true", repo.all[service.SettingKeyOpenAIStrictSchedulerEnabled])
	require.Contains(t, recorder.Body.String(), `"openai_strict_scheduler_enabled":true`)
}

func TestSettingHandler_UpdateSettings_PreservesOpenAIStrictRetrySettingsWhenFieldsOmitted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &settingHandlerRepoStub{
		all: map[string]string{
			service.SettingKeyRegistrationEnabled:          "true",
			service.SettingKeyOpenAIStrictSchedulerEnabled: "true",
			"openai_strict_retry_enabled":                  "true",
			"openai_strict_retry_count":                    "4",
		},
	}
	settingService := service.NewSettingService(repo, &config.Config{})
	handler := NewSettingHandler(settingService, nil, nil, nil, nil, nil)

	router := gin.New()
	router.PUT("/settings", handler.UpdateSettings)

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, "true", repo.all["openai_strict_retry_enabled"])
	require.Equal(t, "4", repo.all["openai_strict_retry_count"])
	require.Contains(t, recorder.Body.String(), `"openai_strict_retry_enabled":true`)
	require.Contains(t, recorder.Body.String(), `"openai_strict_retry_count":4`)
}
