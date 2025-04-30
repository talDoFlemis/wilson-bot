package internal

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"net/http"
	"text/template"
)

//go:embed card_template.json
var cardTemplate []byte

//go:embed broken_card_template.json
var brokenCardTemplate []byte

type GoogleChatProvider interface {
	SendMessage(ctx context.Context, message Message) error
	SendBrokenMessage(ctx context.Context, message BrokenMessage) error
}

type HardcodedGoogleChatProvider struct {
	webhookURL         string
	pearlCardTemplate  *template.Template
	brokenCardTemplate *template.Template
	httpClient         *http.Client
}

type templateData struct {
	Message string
	ID      string
}

type brokenTemplateData struct {
	ID              string
	Name            string
	Motive          string
	TimeSinceBroken string
	DayOfBreakage   string
}

var (
	_ GoogleChatProvider = (*HardcodedGoogleChatProvider)(nil)
)

func NewHardcodedGoogleChatProvider(webhookURL string) (*HardcodedGoogleChatProvider, error) {
	tmpl, err := template.New("email_body.tmpl.xml").Parse(string(cardTemplate))
	if err != nil {
		return nil, err
	}

	brokenTmpl, err := template.New("broken.tmpl.json").Parse(string(brokenCardTemplate))
	if err != nil {
		slog.Error("failed to parse broken card template", slog.Any("error", err))
		return nil, err
	}

	return &HardcodedGoogleChatProvider{
		webhookURL:         webhookURL,
		pearlCardTemplate:  tmpl,
		brokenCardTemplate: brokenTmpl,
		httpClient:         &http.Client{},
	}, nil
}

func (h *HardcodedGoogleChatProvider) SendMessage(ctx context.Context, message Message) error {
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
func (h *HardcodedGoogleChatProvider) SendBrokenMessage(ctx context.Context, message BrokenMessage) error {
	data := brokenTemplateData{
		ID:              message.Id,
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
