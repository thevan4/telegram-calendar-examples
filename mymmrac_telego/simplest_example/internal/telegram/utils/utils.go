package utils

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/constants"
)

// GetChatID ...
func GetChatID(messageOrCallbackQueryOrUpdate interface{}) int64 {
	switch morqu := messageOrCallbackQueryOrUpdate.(type) {
	case telego.Message:
		return morqu.Chat.ID
	case telego.CallbackQuery:
		return morqu.Message.Chat.ID
	case telego.Update:
		return morqu.Message.Chat.ID
	default:
		return 0
	}
}

// GiveMainMenu ...
func GiveMainMenu(text string) *telego.InlineKeyboardMarkup {
	if text == "" {
		return tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton(constants.CallbackCalendarName).WithCallbackData(constants.CallbackCalendar),
			),
		)
	}

	return tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton(text).WithCallbackData(constants.CallbackCalendar),
		),
	)
}
