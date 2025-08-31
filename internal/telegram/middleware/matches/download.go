package matches

import (
	"github.com/go-telegram/bot/models"
)

func MatchUpdate(update *models.Update) bool {

	if update.Message != nil {
		return update.Message.Document != nil
	}

	return false
}
