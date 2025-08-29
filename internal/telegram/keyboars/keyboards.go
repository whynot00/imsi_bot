package keyboars

import "github.com/go-telegram/bot/models"

func URLToMap(url string) *models.InlineKeyboardMarkup {

	var OnMap = models.InlineKeyboardButton{
		Text: "Посмотреть на карте",
		URL:  url,
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				OnMap,
			},
		},
	}
}
