package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	fsm "github.com/whynot00/go-telegram-fsm"
)

func (h *Handler) CancelButton(ctx context.Context, b *bot.Bot, update *models.Update) {

	machine := fsm.FromContext(ctx)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "Отменено.",
		ReplyMarkup: nil,
	})

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	machine.Finish(ctx)
}
