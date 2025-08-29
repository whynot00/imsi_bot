package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/telegram/formatter"
	"github.com/whynot00/imsi_bot/internal/telegram/keyboars"
)

func (h *Handler) FetchObservation(ctx context.Context, b *bot.Bot, update *models.Update) {
	var err error

	var obs []observation.Observation

	switch {
	case strings.HasPrefix(update.Message.Text, "250"):
		obs, err = h.obs.GetByIMSI(ctx, update.Message.Text)
		if err != nil {
			messageError(ctx, b, errorx.ReqError{
				UserID:  update.Message.From.ID,
				Request: update.Message.Text,
				Err:     err,
			})
		}

	default:
		obs, err = h.obs.GetByIMEI(ctx, update.Message.Text)
		if err != nil {
			messageError(ctx, b, errorx.ReqError{
				UserID:  update.Message.From.ID,
				Request: update.Message.Text,
				Err:     err,
			})
		}
	}

	if obs == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.From.ID,
			Text:      formatter.NoMatchesMessage(),
			ParseMode: models.ParseModeMarkdown,
		})

		return
	}

	for _, o := range obs {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.From.ID,
			Text:        formatter.ObservationMessage(&o),
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: keyboars.URLToMap(fmt.Sprintf("https://yandex.ru/maps/?pt=%.6f,%.6f&z=14&l=map", o.LON, o.LAT)),
		})

	}

}

func (h *Handler) CreateUser(ctx context.Context, b *bot.Bot, update *models.Update) {

	raw := strings.TrimPrefix(update.Message.Text, "/new ")
	id, err := strconv.ParseInt(raw, 10, 64)

	if err != nil {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.From.ID,
			Text:      formatter.InvalidID(raw),
			ParseMode: models.ParseModeMarkdown,
		})

		return
	}

	if err := h.whl.Create(ctx, id); err != nil {

		if errors.Is(err, errorx.ErrUserIsExists) {

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.From.ID,
				Text:      formatter.UserAlreadyExistsByID(id),
				ParseMode: models.ParseModeMarkdown,
			})

		}

		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.From.ID,
		Text:      formatter.UserAddedByID(raw),
		ParseMode: models.ParseModeMarkdown,
	})

}

func (h *Handler) DownloadUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {

	// filepath := "data/" + update.Message.Document.FileUniqueID + ".csv"
	// if err := downloadFile(ctx, b, update.Message.Document.FileID, filepath); err != nil {
	// 	messageError(ctx, b, errorx.ReqError{
	// 		UserID:  update.Message.From.ID,
	// 		Request: "Загрузка файла",
	// 		Err:     err,
	// 	})

	// 	return
	// }

	// datapump.Pump(filepath, h.db)

	// os.Remove(filepath)
}

func downloadFile(ctx context.Context, b *bot.Bot, fileID, destPath string) error {

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return fmt.Errorf("GetFile: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), file.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}

func messageError(ctx context.Context, b *bot.Bot, err errorx.ReqError) {

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
