package main

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

type Server struct {
	messageStorer MessageStorer
	echoServer    *echo.Echo
}

func NewServer(cfg HTTPConfig, messageStorer MessageStorer) *Server {
	e := echo.New()

	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.Recover())

	server := &Server{
		messageStorer: messageStorer,
		echoServer:    e,
	}

	api := e.Group(cfg.Prefix)
	messagesRouter := api.Group("/messages")

	messagesRouter.GET("/", server.GetAllMessages)
	messagesRouter.GET("/:id", server.GetMessageById)

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

func (s *Server) Start(addr string) error {
	return s.echoServer.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echoServer.Shutdown(ctx)
}
