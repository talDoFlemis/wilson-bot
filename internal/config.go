package internal

import (
	"context"
	_ "embed"
	"log/slog"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

const (
	PREFIX = "WILSONBOT__"
)

var k = koanf.New(".")

//go:embed base_config.toml
var baseConfig []byte

type HTTPConfig struct {
	Port       string `koanf:"port"`
	Prefix     string `koanf:"prefix"`
	Host       string `koanf:"host"`
	EnableSend bool   `koanf:"enable_send"`
}

type CronConfig struct {
	Enabled bool `koanf:"enabled"`
	CronString string `koanf:"cron_string"`
}

type GoogleChatConfig struct {
	WebhookURL string `koanf:"webhook_url"`
}

type DiscordWebhookConfig struct {
	WebhookURL string `koanf:"webhook_url"`
}

type Config struct {
	HTTPConfig       HTTPConfig       `koanf:"http"`
	CronConfig       CronConfig       `koanf:"cron"`
	GoogleChatConfig GoogleChatConfig `koanf:"google_chat"`
	DiscordWebhookConfig DiscordWebhookConfig `koanf:"discord_webhook"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	err := k.Load(rawbytes.Provider(baseConfig), toml.Parser())
	if err != nil {
		slog.ErrorContext(ctx, "failed to load base config", slog.Any("error", err))
		return nil, err
	}

	err = k.Load(env.Provider(PREFIX, ".", func(s string) string {
		lowerCased := strings.ToLower(s)
		replacer := strings.NewReplacer(strings.ToLower(PREFIX), "", "__", ".")
		return replacer.Replace(lowerCased)
	}), nil)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load envs from env variables", slog.Any("error", err))
		return nil, err
	}

	var out Config

	err = k.Unmarshal("", &out)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshall config struct", slog.Any("error", err))
		return nil, err
	}

	slog.InfoContext(ctx, "config loaded")

	return &out, nil
}
