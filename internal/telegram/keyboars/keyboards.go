package keyboars

import "github.com/go-telegram/bot/models"

var Cancel = models.InlineKeyboardButton{
	Text:         "Отмена",
	CallbackData: "cancel_button",
}

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

func CancelButton() *models.InlineKeyboardMarkup {

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				Cancel,
			},
		},
	}
}
