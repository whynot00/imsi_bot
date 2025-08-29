package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
)

func (m *Middleware) Whitelist(next bot.HandlerFunc) bot.HandlerFunc {

	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

		user, err := m.whl.Touch(ctx, update.Message.From.ID)
		if err != nil {
			if errors.Is(err, errorx.ErrNoRows) {
				return
			}

			fmt.Println(err)
			return
		}

		if user.Username != update.Message.From.Username {
			m.whl.Update(ctx, &whitelist.User{
				Username:   update.Message.From.Username,
				TelegramID: update.Message.From.ID,
			})
		}

		next(ctx, bot, update)
	}
}
