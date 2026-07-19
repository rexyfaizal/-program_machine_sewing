package bot

import (
	"context"
	"errors"
	"log"
	"time"

	"telegram_notif/handler"
	"telegram_notif/models"
)

type UpdateClient interface {
	GetUpdates(
		ctx context.Context,
		offset int64,
	) ([]models.TelegramUpdate, error)
}

type Poller struct {
	client  UpdateClient
	handler *handler.TelegramHandler
}

func NewPoller(
	client UpdateClient,
	handler *handler.TelegramHandler,
) *Poller {
	return &Poller{
		client:  client,
		handler: handler,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	var offset int64

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		requestCtx, cancel := context.WithTimeout(ctx, 65*time.Second)
		updates, err := p.client.GetUpdates(requestCtx, offset)
		cancel()

		if err != nil {
			if errors.Is(err, context.Canceled) && ctx.Err() != nil {
				return nil
			}

			log.Printf("Gagal mengambil update Telegram: %v", err)

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			p.handler.Handle(ctx, update)
		}
	}
}
