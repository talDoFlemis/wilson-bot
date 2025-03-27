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

type GoogleChatProvider interface {
	SendMessage(ctx context.Context, message Message) error
}

type HardcodedGoogleChatProvider struct {
	webhookURL   string
	cardTemplate *template.Template
	httpClient   *http.Client
}

type templateData struct {
	Message string
	ID      string
}

var (
	_ GoogleChatProvider = (*HardcodedGoogleChatProvider)(nil)
)

func NewHardcodedGoogleChatProvider(webhookURL string) (*HardcodedGoogleChatProvider, error) {
	tmpl, err := template.New("email_body.tmpl.xml").Parse(string(cardTemplate))
	if err != nil {
		return nil, err
	}

	return &HardcodedGoogleChatProvider{
		webhookURL:   webhookURL,
		cardTemplate: tmpl,
		httpClient:   &http.Client{},
	}, nil
}

func (h *HardcodedGoogleChatProvider) SendMessage(ctx context.Context, message Message) error {
	data := templateData{
		Message: message.Message,
	}

	var buf bytes.Buffer

	err := h.cardTemplate.Execute(&buf, data)
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
