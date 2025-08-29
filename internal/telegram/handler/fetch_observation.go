package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/telegram/formatter"
	"github.com/whynot00/imsi_bot/internal/telegram/handler/utils"
	"github.com/whynot00/imsi_bot/internal/telegram/keyboars"
)

// FetchObservation обрабатывает входящее сообщение, содержащее IMEI или IMSI.
//
// Если строка соответствует регулярному выражению `^\d{14,15}$`,
// выполняется поиск записи в БД по этому значению.
//
// При успешном нахождении возвращает:
//   - стандарт
//   - уровень сигнала
//   - дата/время
//   - координаты
//   - оператор (опц.)
//   - IMSI       (опц.)
//   - IMEI       (опц.)
func (h *Handler) FetchObservation(ctx context.Context, b *bot.Bot, update *models.Update) {
	var err error

	var obs []observation.Observation

	// * Определяем IMSI или IMEI
	switch {
	case strings.HasPrefix(update.Message.Text, "250"):

		// * Запрашиваем по IMSI
		obs, err = h.obs.GetByIMSI(ctx, update.Message.Text)
		if err != nil {
			utils.MessageError(ctx, b, errorx.ReqError{
				UserID:  update.Message.From.ID,
				Request: update.Message.Text,
				Err:     err,
			})
		}

	default:

		// * Запрашиваем по IMEI
		obs, err = h.obs.GetByIMEI(ctx, update.Message.Text)

		if err != nil {
			utils.MessageError(ctx, b, errorx.ReqError{
				UserID:  update.Message.From.ID,
				Request: update.Message.Text,
				Err:     err,
			})
		}
	}

	// * Если пусто то возвращаем сообщение "Совпадений нет"
	if obs == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.From.ID,
			Text:      formatter.NoMatchesMessage(),
			ParseMode: models.ParseModeMarkdown,
		})

		return
	}

	// * Если совпадения есть, то выводим все
	for _, o := range obs {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.From.ID,
			Text:        formatter.ObservationMessage(&o),
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: keyboars.URLToMap(fmt.Sprintf("https://yandex.ru/maps/?pt=%.6f,%.6f&z=14&l=map", o.LON, o.LAT)),
		})

	}

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
