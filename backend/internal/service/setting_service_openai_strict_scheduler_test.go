//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type strictSchedulerSettingsRepoStub struct {
	all map[string]string
}

func (s *strictSchedulerSettingsRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *strictSchedulerSettingsRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.all == nil {
		return "", ErrSettingNotFound
	}
	value, ok := s.all[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (s *strictSchedulerSettingsRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *strictSchedulerSettingsRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *strictSchedulerSettingsRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *strictSchedulerSettingsRepoStub) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.all))
	for key, value := range s.all {
		out[key] = value
	}
	return out, nil
}

func (s *strictSchedulerSettingsRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestSettingService_GetAllSettings_MissingOpenAIStrictSchedulerFallsBackToConfig(t *testing.T) {
	repo := &strictSchedulerSettingsRepoStub{
		all: map[string]string{
			SettingKeyRegistrationEnabled: "true",
		},
	}
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.SchedulerMode = "strict_priority_fallback"

	svc := NewSettingService(repo, cfg)
	settings, err := svc.GetAllSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.OpenAIStrictSchedulerEnabled)
}

func TestSettingService_GetAllSettings_ExplicitOpenAIStrictSchedulerOverridesConfigFallback(t *testing.T) {
	repo := &strictSchedulerSettingsRepoStub{
		all: map[string]string{
			SettingKeyRegistrationEnabled:          "true",
			SettingKeyOpenAIStrictSchedulerEnabled: "false",
		},
	}
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.SchedulerMode = "strict_priority_fallback"

	svc := NewSettingService(repo, cfg)
	settings, err := svc.GetAllSettings(context.Background())
	require.NoError(t, err)
	require.False(t, settings.OpenAIStrictSchedulerEnabled)
}
