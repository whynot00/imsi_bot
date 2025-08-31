package utils

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/telegram/formatter"
)

func MessageError(ctx context.Context, b *bot.Bot, err errorx.ReqError) {

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: err.UserID,
		Text:   formatter.InternalError(),
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    "981397216",
		Text:      formatter.InternalErrorAdmin(err),
		ParseMode: models.ParseModeHTML,
	})

}
