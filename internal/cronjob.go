package internal

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// MessageCronJob handles scheduled message sending tasks
type MessageCronJob struct {
	messageStorer      MessageStorer
	googleChatProvider GoogleChatProvider
	scheduler          gocron.Scheduler
	cronString         string
	enabled            bool
}

// NewMessageCronJob creates a new cron job service for scheduled messages
func NewMessageCronJob(
	cfg CronConfig,
	messageStorer MessageStorer,
	googleChatProvider GoogleChatProvider,
) (*MessageCronJob, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("failed to create cron scheduler", slog.Any("error", err))
		return nil, err
	}

	return &MessageCronJob{
		messageStorer:      messageStorer,
		googleChatProvider: googleChatProvider,
		enabled:            cfg.Enabled,
		cronString:         cfg.CronString,
		scheduler:          scheduler,
	}, nil
}

// Start begins the cron scheduler
func (c *MessageCronJob) Start(ctx context.Context) error {
	if !c.enabled {
		slog.InfoContext(ctx, "cron jobs are disabled, not starting scheduler")
		return nil
	}

	job, err := c.scheduler.NewJob(
		gocron.CronJob(c.cronString, false),
		gocron.NewTask(func() {
			c.sendDailyMessage(context.Background())
		}),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to schedule daily message job", slog.Any("error", err))
		return err
	}

	c.scheduler.Start()

	slog.InfoContext(ctx, "cron scheduler started, message will be sent at", slog.String("cron_string", c.cronString), slog.Any("job_id", job.ID()))

	return nil
}

// Stop halts the cron scheduler
func (c *MessageCronJob) Stop(ctx context.Context) {
	if c.scheduler != nil {
		err := c.scheduler.Shutdown()
		if err != nil {
			slog.ErrorContext(ctx, "failed to stop cron scheduler", slog.Any("error", err))
		} else {
			slog.InfoContext(ctx, "cron scheduler stopped successfully")
		}
	}
}

// sendDailyMessage sends a random message from the repository to Google Chat
func (c *MessageCronJob) sendDailyMessage(ctx context.Context) {
	slog.InfoContext(ctx, "executing daily message job")

	messages, err := c.messageStorer.GetAllMessages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get messages", slog.Any("error", err))
		return
	}

	if len(messages) == 0 {
		slog.ErrorContext(ctx, "no messages available for sending")
		return
	}

	// Select a random message
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(messages))
	randomMessage := messages[randomIndex]

	err = c.googleChatProvider.SendMessage(ctx, randomMessage)
	if err != nil {
		slog.ErrorContext(ctx, "failed to send message", slog.Any("error", err))
		return
	}

	slog.InfoContext(ctx, "daily message sent successfully",
		slog.String("message_id", randomMessage.Id),
		slog.String("message", randomMessage.Message))
}
