package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	fsm "github.com/whynot00/go-telegram-fsm"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/states"
	"github.com/whynot00/imsi_bot/internal/telegram/handler/utils"
	"github.com/whynot00/imsi_bot/internal/telegram/keyboars"
)

func (h *Handler) DownloadUpdateCommand(ctx context.Context, b *bot.Bot, update *models.Update) {

	machine := fsm.FromContext(ctx)

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        "Отправьте файл с расширением `.csv`",
		ReplyMarkup: keyboars.CancelButton(),
	})
	if err != nil {
		fmt.Println(err)
	}

	machine.Set(ctx, update.Message.From.ID, UpdateDataKey, msg.ID)

	machine.Transition(ctx, states.UploadFileState)

}

func (h *Handler) DownloadUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	machine := fsm.FromContext(ctx)

	defer machine.Finish(ctx)

	filepath := "data/" + update.Message.Document.FileUniqueID + ".csv"
	if err := utils.DownloadFile(ctx, b, update.Message.Document.FileID, filepath); err != nil {
		utils.MessageError(ctx, b, errorx.ReqError{
			UserID:  update.Message.From.ID,
			Request: "Загрузка файла",
			Err:     err,
		})

		return
	}

	if err := h.pump.Pump(ctx, filepath); err != nil {
		utils.MessageError(ctx, b, errorx.ReqError{
			UserID:  update.Message.From.ID,
			Request: "Загрузка файла",
			Err:     err,
		})
	}

	os.Remove(filepath)
}
