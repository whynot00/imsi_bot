package utils

import "github.com/go-telegram/bot/models"

func ExtractUserID(u *models.Update) int64 {
	switch {
	case u.Message != nil && u.Message.From != nil:
		return u.Message.From.ID
	case u.EditedMessage != nil && u.EditedMessage.From != nil:
		return u.EditedMessage.From.ID
	case u.BusinessMessage != nil && u.BusinessMessage.From != nil:
		return u.BusinessMessage.From.ID
	case u.EditedBusinessMessage != nil && u.EditedBusinessMessage.From != nil:
		return u.EditedBusinessMessage.From.ID
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From.ID
	case u.InlineQuery != nil && u.InlineQuery.From != nil:
		return u.InlineQuery.From.ID
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From.ID
	case u.ShippingQuery != nil && u.ShippingQuery.From != nil:
		return u.ShippingQuery.From.ID
	case u.PreCheckoutQuery != nil && u.PreCheckoutQuery.From != nil:
		return u.PreCheckoutQuery.From.ID
	case u.PurchasedPaidMedia != nil:
		return u.PurchasedPaidMedia.From.ID
	case u.ChatMember != nil:
		return u.ChatMember.From.ID
	case u.MyChatMember != nil:
		return u.MyChatMember.From.ID
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.From.ID
	case u.PollAnswer != nil && u.PollAnswer.User != nil:
		return u.PollAnswer.User.ID
	case u.MessageReaction != nil && u.MessageReaction.User != nil:
		return u.MessageReaction.User.ID
	}
	return 0
}

func ExtractUsername(u *models.Update) string {
	switch {
	case u.Message != nil && u.Message.From != nil:
		return u.Message.From.Username
	case u.EditedMessage != nil && u.EditedMessage.From != nil:
		return u.EditedMessage.From.Username
	case u.BusinessMessage != nil && u.BusinessMessage.From != nil:
		return u.BusinessMessage.From.Username
	case u.EditedBusinessMessage != nil && u.EditedBusinessMessage.From != nil:
		return u.EditedBusinessMessage.From.Username
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From.Username
	case u.InlineQuery != nil && u.InlineQuery.From != nil:
		return u.InlineQuery.From.Username
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From.Username
	case u.ShippingQuery != nil && u.ShippingQuery.From != nil:
		return u.ShippingQuery.From.Username
	case u.PreCheckoutQuery != nil && u.PreCheckoutQuery.From != nil:
		return u.PreCheckoutQuery.From.Username
	case u.PurchasedPaidMedia != nil:
		return u.PurchasedPaidMedia.From.Username
	case u.ChatMember != nil:
		return u.ChatMember.From.Username
	case u.MyChatMember != nil:
		return u.MyChatMember.From.Username
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.From.Username
	case u.PollAnswer != nil && u.PollAnswer.User != nil:
		return u.PollAnswer.User.Username
	case u.MessageReaction != nil && u.MessageReaction.User != nil:
		return u.MessageReaction.User.Username
	}
	return ""
}
