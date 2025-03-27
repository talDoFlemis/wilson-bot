package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock MessageStorer
type MockMessageStorer struct {
	mock.Mock
}

func (m *MockMessageStorer) GetAllMessages(ctx context.Context) ([]Message, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Message), args.Error(1)
}

func (m *MockMessageStorer) GetMessageByID(ctx context.Context, id string) (*Message, error) {
	args := m.Called(ctx, id)
	if msg := args.Get(0); msg != nil {
		return msg.(*Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestGetAllMessages(t *testing.T) {
	e := echo.New()
	mockStore := new(MockMessageStorer)

	expectedMessages := []Message{
		{Id: "1", Message: "Hello"},
		{Id: "2", Message: "World"},
	}
	mockStore.On("GetAllMessages", mock.Anything).Return(expectedMessages, nil)

	server := &Server{messageStorer: mockStore, echoServer: e}

	req := httptest.NewRequest(http.MethodGet, "/messages/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.GetAllMessages(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestGetMessageByIdSuccess(t *testing.T) {
	e := echo.New()
	mockStore := new(MockMessageStorer)

	expectedMessage := &Message{Id: "1", Message: "Hello"}
	mockStore.On("GetMessageByID", mock.Anything, "1").Return(expectedMessage, nil)

	server := &Server{messageStorer: mockStore, echoServer: e}

	req := httptest.NewRequest(http.MethodGet, "/messages/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := server.GetMessageById(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestGetMessageByIdNotFound(t *testing.T) {
	e := echo.New()
	mockStore := new(MockMessageStorer)

	mockStore.On("GetMessageByID", mock.Anything, "999").Return(nil, ErrMessageNotFound)

	server := &Server{messageStorer: mockStore, echoServer: e}

	req := httptest.NewRequest(http.MethodGet, "/messages/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := server.GetMessageById(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockStore.AssertExpectations(t)
}

func TestNewServer(t *testing.T) {
	mockStore := new(MockMessageStorer)
	cfg := HTTPConfig{Prefix: "/api"}

	server := NewServer(cfg, mockStore)

	assert.NotNil(t, server)
	assert.NotNil(t, server.echoServer)
}
