package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	fsm "github.com/whynot00/go-telegram-fsm"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/states"
	"github.com/whynot00/imsi_bot/internal/telegram/formatter"
	"github.com/whynot00/imsi_bot/internal/telegram/keyboars"
)

// CreateUserCommand реакция на комманду /add_user
func (h *Handler) CreateUserCommand(ctx context.Context, b *bot.Bot, update *models.Update) {

	machine := fsm.FromContext(ctx)

	msg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        "Введите ID пользователя для регистрации.",
		ReplyMarkup: keyboars.CancelButton(),
	})

	machine.Set(ctx, update.Message.From.ID, CreateNewUserMessageKey, msg.ID)

	machine.Transition(ctx, states.CreateUserState)
}

// CreateUser создание нового пользователя
func (h *Handler) CreateUser(ctx context.Context, b *bot.Bot, update *models.Update) {

	// * достаем FSM из контекста
	machine := fsm.FromContext(ctx)

	// * по окончании работы этого метода обязаны обнулить state
	defer machine.Finish(ctx)

	// * достаем ID сообщения на которое ответил пользовател
	var messageID int
	if id, ok := machine.Get(ctx, update.Message.From.ID, CreateNewUserMessageKey); ok {
		messageID = id.(int)
	}

	id, err := strconv.ParseInt(update.Message.Text, 10, 64)
	if err != nil {

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.Message.From.ID,
			MessageID:   messageID,
			Text:        formatter.InvalidID(update.Message.Text),
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: nil,
		})

		return
	}

	// * записываем нового пользователя в БД
	if err := h.whl.Create(ctx, id); err != nil {

		if errors.Is(err, errorx.ErrUserIsExists) {

			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      update.Message.From.ID,
				MessageID:   messageID,
				Text:        formatter.UserAlreadyExistsByID(id),
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: nil,
			})

		}

		return
	}

	// * при успехе уведомляем пользователя
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.Message.From.ID,
		MessageID:   messageID,
		Text:        formatter.UserAddedByID(update.Message.Text),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: nil,
	})

}
