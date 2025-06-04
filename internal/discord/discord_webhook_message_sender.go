package discord

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	_ "embed"

	"github.com/taldoflemis/wilson-bot/internal"
)

//go:embed thanking_card_template.json
var cardTemplate []byte

//go:embed broken_card_template.json
var brokenCardTemplate []byte

type DiscordWebhookMessageSender struct {
	webhookURL         string
	pearlCardTemplate  *template.Template
	brokenCardTemplate *template.Template
	httpClient         *http.Client
}

type templateData struct {
	Message string
}

type brokenTemplateData struct {
	Name            string
	Motive          string
	TimeSinceBroken string
	DayOfBreakage   string
}

var (
	_ internal.MessageSender = (*DiscordWebhookMessageSender)(nil)
)

func NewDiscordWebhookMessageSender(webhookURL string) (*DiscordWebhookMessageSender, error) {
	tmpl, err := template.New("email_body.tmpl.xml").Parse(string(cardTemplate))
	if err != nil {
		return nil, err
	}

	brokenTmpl, err := template.New("broken.tmpl.json").Parse(string(brokenCardTemplate))
	if err != nil {
		slog.Error("failed to parse broken card template", slog.Any("error", err))
		return nil, err
	}

	return &DiscordWebhookMessageSender{
		webhookURL:         webhookURL,
		pearlCardTemplate:  tmpl,
		brokenCardTemplate: brokenTmpl,
		httpClient:         &http.Client{},
	}, nil
}

func (h *DiscordWebhookMessageSender) SendMessage(ctx context.Context, message internal.Message) error {
	data := templateData{
		Message: message.Message,
	}

	var buf bytes.Buffer

	err := h.pearlCardTemplate.Execute(&buf, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to execute template", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.webhookURL, &buf)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create request", slog.Any("error", err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to send request", slog.Any("error", err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "unexpected status code", slog.Any("status_code", resp.StatusCode))
		return errors.New("unexpected status code")
	}

	return nil
}

// SendBrokenMessage implements GoogleChatProvider.
func (h *DiscordWebhookMessageSender) SendBrokenMessage(ctx context.Context, message internal.BrokenMessage) error {
	data := brokenTemplateData{
		Name:            message.Name,
		Motive:          message.Motive,
		TimeSinceBroken: message.TimeSinceBroken,
		DayOfBreakage:   message.DayOfBreakage,
	}

	var buf bytes.Buffer

	err := h.brokenCardTemplate.Execute(&buf, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to execute template", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.webhookURL, &buf)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create request", slog.Any("error", err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to send request", slog.Any("error", err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "unexpected status code", slog.Any("status_code", resp.StatusCode))
		return errors.New("unexpected status code")
	}

	return nil
}
