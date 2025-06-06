package internal

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

type Server struct {
	messageStorer MessageStorer
	messageSender MessageSender
	sendMessages  bool
	echoServer    *echo.Echo
}

func NewServer(
	cfg HTTPConfig,
	messageStorer MessageStorer,
	messageSender MessageSender,
) *Server {
	e := echo.New()

	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.Recover())

	server := &Server{
		messageStorer: messageStorer,
		messageSender: messageSender,
		echoServer:    e,
		sendMessages:  cfg.EnableSend,
	}

	api := e.Group(cfg.Prefix)
	api.GET("/healthz", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	messagesRouter := api.Group("/messages")

	messagesRouter.GET("/", server.GetAllMessages)
	messagesRouter.GET("/:id", server.GetMessageById)
	messagesRouter.POST("/", server.SendMessage)
	messagesRouter.POST("/:id", server.SendMessageById)

	webhookRouter := api.Group("/webhook")
	webhookRouter.POST("/broken", server.SendBrokenMessageWebhook)

	return server
}

func (s *Server) GetAllMessages(c echo.Context) error {
	messages, err := s.messageStorer.GetAllMessages(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, messages)
}

func (s *Server) GetMessageById(c echo.Context) error {
	id := c.Param("id")

	message, err := s.messageStorer.GetMessageByID(c.Request().Context(), id)
	if err == nil {
		return c.JSON(200, message)
	}

	if err == ErrMessageNotFound {
		return c.JSON(404, map[string]string{"error": err.Error()})
	}

	return c.JSON(500, map[string]string{"error": err.Error()})
}

func (s *Server) SendMessageById(c echo.Context) error {
	id := c.Param("id")

	if !s.sendMessages {
		return c.JSON(403, map[string]string{"error": "sending messages is disabled"})
	}

	message, err := s.messageStorer.GetMessageByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	err = s.messageSender.SendMessage(c.Request().Context(), *message)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "message sent"})
}

func (s *Server) SendMessage(c echo.Context) error {
	if !s.sendMessages {
		return c.JSON(403, map[string]string{"error": "sending messages is disabled"})
	}

	messages, err := s.messageStorer.GetAllMessages(c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	if len(messages) == 0 {
		return c.JSON(404, map[string]string{"error": "no messages available"})
	}

	// Seed random generator (only needed once in main function)
	rand.Seed(time.Now().UnixNano())

	// Select a random message
	randomIndex := rand.Intn(len(messages))
	randomMessage := messages[randomIndex]

	err = s.messageSender.SendMessage(c.Request().Context(), randomMessage)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "message sent"})
}

func (s *Server) SendBrokenMessageWebhook(c echo.Context) error {
	if !s.sendMessages {
		return c.JSON(403, map[string]string{"error": "sending messages is disabled"})
	}

	var brokenMessage BrokenMessage
	if err := c.Bind(&brokenMessage); err != nil {
		return c.JSON(400, map[string]string{"error": "invalid request"})
	}

	err := s.messageSender.SendBrokenMessage(c.Request().Context(), brokenMessage)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "broken message sent"})
}

func (s *Server) Start(addr string) error {
	return s.echoServer.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echoServer.Shutdown(ctx)
}
