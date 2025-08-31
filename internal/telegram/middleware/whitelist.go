package middleware

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
	"github.com/whynot00/imsi_bot/internal/telegram/middleware/utils"
)

func (m *Middleware) Whitelist(next bot.HandlerFunc) bot.HandlerFunc {

	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

		userID := utils.ExtractUserID(update)
		username := utils.ExtractUsername(update)

		user, err := m.whl.Touch(ctx, userID)
		if err != nil {
			if errors.Is(err, errorx.ErrNoRows) {
				return
			}

			return
		}

		if user.Username != username {
			m.whl.Update(ctx, &whitelist.User{
				Username:   username,
				TelegramID: userID,
			})
		}

		next(ctx, bot, update)
	}
}
