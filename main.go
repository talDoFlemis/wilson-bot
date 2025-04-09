package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/taldoflemis/wilson-bot/internal"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	retcode := 0

	defer func() {
		os.Exit(retcode)
	}()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	cfg, err := internal.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", slog.Any("error", err))
		return
	}

	messages, err := internal.GetMessages(ctx, internal.RawMessages)

	if err != nil {
		slog.InfoContext(ctx, "Failed to get messages", slog.Any("error", err))
		retcode = 1
		return
	}

	hardcodedGoogleChatProvider, err := internal.NewHardcodedGoogleChatProvider(
		cfg.GoogleChatConfig.WebhookURL,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create google chat provider", slog.Any("error", err))
		retcode = 1
		return
	}

	dumpMessageStorer := internal.NewMessageStorer(messages)

	messageCronJob, err := internal.NewMessageCronJob(cfg.CronConfig, dumpMessageStorer, hardcodedGoogleChatProvider)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create message cron job", slog.Any("error", err))
		retcode = 1
		return
	}

	err = messageCronJob.Start(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to start message cron job", slog.Any("error", err))
		retcode = 1
		return
	}

	server := internal.NewServer(cfg.HTTPConfig, dumpMessageStorer, hardcodedGoogleChatProvider)
	errChan := make(chan error)

	go func() {
		err = server.Start(cfg.HTTPConfig.Host + ":" + cfg.HTTPConfig.Port)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err = <-errChan:
		slog.ErrorContext(ctx, "error when running server", slog.Any("err", err))
		retcode = 1
	case <-ctx.Done():
		// Wait for first Signal arrives
	}

	err = server.Shutdown(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to shutdown gracefully server", slog.Any("error", err))
		retcode = 1
		return
	}

	// Stop the cron job gracefully
	messageCronJob.Stop(ctx)
}
