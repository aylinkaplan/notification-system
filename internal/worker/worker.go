package worker

import (
	"context"
	"time"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/aylinkaplan/notification-system/internal/provider"
	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type NotificationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, providerMsgID, errMsg *string) error
	IncrementRetry(ctx context.Context, id uuid.UUID, errMsg string) error
}

type Limiter interface {
	Wait(ctx context.Context, key string) error
}

type Worker struct {
	queue    queue.Queue
	repo     NotificationRepository
	provider *provider.WebhookProvider
	limiter  Limiter
	webhook  string
	channels []string
}

func NewWorker(q queue.Queue, repo NotificationRepository, prov *provider.WebhookProvider, limiter Limiter, webhookUUID string) *Worker {
	return &Worker{
		queue:    q,
		repo:     repo,
		provider: prov,
		limiter:  limiter,
		webhook:  webhookUUID,
		channels: []string{"sms", "email", "push"},
	}
}

func (w *Worker) Run(ctx context.Context) {
	for _, ch := range w.channels {
		go w.processChannel(ctx, ch)
	}
}

func (w *Worker) processChannel(ctx context.Context, channel string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			item, err := w.queue.Pop(ctx, channel, 2*time.Second)
			if err != nil {
				log.Error().Err(err).Str("channel", channel).Msg("queue pop failed")
				continue
			}
			if item == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if err := w.limiter.Wait(ctx, "channel:"+channel); err != nil {
				log.Error().Err(err).Msg("rate limit wait failed")
				w.queue.Push(ctx, channel, item.NotificationID, item.Priority)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			w.process(ctx, item)
		}
	}
}

func (w *Worker) process(ctx context.Context, item *queue.QueuedItem) {
	n, err := w.repo.GetByID(ctx, item.NotificationID)
	if err != nil || n == nil {
		log.Error().Err(err).Str("id", item.NotificationID.String()).Msg("notification not found")
		return
	}
	if n.Status != domain.StatusPending {
		return
	}

	// Mark processing
	w.repo.UpdateStatus(ctx, n.ID, domain.StatusProcessing, nil, nil)

	resp, err := w.provider.Send(ctx, w.webhook, n.Recipient, string(n.Channel), n.Content)
	if err != nil {
		n.RetryCount++
		errMsg := err.Error()
		if n.RetryCount >= n.MaxRetries {
			w.repo.UpdateStatus(ctx, n.ID, domain.StatusFailed, nil, &errMsg)
			log.Error().Err(err).Str("id", n.ID.String()).Msg("delivery failed after max retries")
		} else {
			w.repo.IncrementRetry(ctx, n.ID, errMsg)
			w.repo.UpdateStatus(ctx, n.ID, domain.StatusPending, nil, nil)
			time.Sleep(backoff(n.RetryCount))
			w.queue.Push(ctx, string(n.Channel), n.ID, string(n.Priority))
		}
		return
	}

	msgID := resp.MessageID
	w.repo.UpdateStatus(ctx, n.ID, domain.StatusDelivered, &msgID, nil)
	log.Info().Str("id", n.ID.String()).Str("provider_msg_id", msgID).Msg("delivered")
}

func backoff(retry int) time.Duration {
	return Backoff(retry)
}

// Backoff returns the delay for the given retry count (used by backoff and tests).
func Backoff(retry int) time.Duration {
	d := time.Duration(1<<uint(retry)) * time.Second
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}
