package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMessages(t *testing.T) {
	ctx := context.Background()

	// Valid JSON
	jsonData := []byte(
		`[{"id": "1", "message": "Hello", "sentiment": "positive", "tags": ["greeting"]}]`,
	)
	messages, err := GetMessages(ctx, jsonData)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "1", messages[0].Id)
	assert.Equal(t, "Hello", messages[0].Message)

	// Invalid JSON
	invalidJSON := []byte(`{"id": "1", "message": "Hello"}`)
	messages, err = GetMessages(ctx, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, messages)
}

func TestDumpMessageStorerGetAllMessages(t *testing.T) {
	ctx := context.Background()
	mockMessages := []Message{
		{Id: "1", Message: "Hello", Sentiment: "positive", Tags: []string{"greeting"}},
		{Id: "2", Message: "Bye", Sentiment: "negative", Tags: []string{"farewell"}},
	}

	storer := NewMessageStorer(mockMessages)
	messages, err := storer.GetAllMessages(ctx)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
}

func TestDumpMessageStorerGetMessageByID(t *testing.T) {
	ctx := context.Background()
	mockMessages := []Message{
		{Id: "1", Message: "Hello", Sentiment: "positive", Tags: []string{"greeting"}},
	}

	storer := NewMessageStorer(mockMessages)

	// Existing message
	message, err := storer.GetMessageByID(ctx, "1")
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "1", message.Id)

	// Non-existing message
	message, err = storer.GetMessageByID(ctx, "999")
	assert.ErrorIs(t, err, ErrMessageNotFound)
	assert.Nil(t, message)
}
