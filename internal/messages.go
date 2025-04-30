package internal

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
)

//go:embed messages.json
var RawMessages []byte

var (
	ErrMessageNotFound = errors.New("message not found")
)

type Message struct {
	Id        string   `json:"id"`
	Message   string   `json:"message"`
	Sentiment string   `json:"sentiment"`
	Tags      []string `json:"tags"`
}

type BrokenMessage struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Motive          string `json:"motive"`
	TimeSinceBroken string `json:"time_since_broken"`
	DayOfBreakage   string `json:"day_of_breakage"`
}

func GetMessages(ctx context.Context, rawMessagesData []byte) ([]Message, error) {
	var messages []Message

	err := json.Unmarshal(rawMessagesData, &messages)
	if err != nil {
		slog.InfoContext(ctx, "Failed to desserialize messages", slog.Any("error", err))
		return nil, err
	}

	return messages, nil
}

type MessageStorer interface {
	GetAllMessages(ctx context.Context) ([]Message, error)
	GetMessageByID(ctx context.Context, id string) (*Message, error)
}

type DumpMessageStorer struct {
	Messages []Message
}

var (
	_ MessageStorer = (*DumpMessageStorer)(nil)
)

func NewMessageStorer(messages []Message) *DumpMessageStorer {
	return &DumpMessageStorer{
		Messages: messages,
	}
}

func (s *DumpMessageStorer) GetAllMessages(ctx context.Context) ([]Message, error) {
	return s.Messages, nil
}

func (s *DumpMessageStorer) GetMessageByID(ctx context.Context, id string) (*Message, error) {
	for _, m := range s.Messages {
		if m.Id == id {
			return &m, nil
		}
	}

	return nil, ErrMessageNotFound
}
