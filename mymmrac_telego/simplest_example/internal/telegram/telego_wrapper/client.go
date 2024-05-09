package telego_wrapper

import (
	"fmt"
	"log"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/calendar_wrapper"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/constants"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/utils"
)

// ClientWrapper ...
type ClientWrapper struct {
	*telego.Bot
}

// MustNewClient ...
func MustNewClient(botToken string, options ...telego.BotOption) *ClientWrapper {
	bot, err := telego.NewBot(botToken, options...)
	if err != nil {
		log.Panicf("at MustNewClient: %v", err)
	}
	return &ClientWrapper{
		Bot: bot,
	}
}

// MustGetLongPollingUpdates ...
func (c *ClientWrapper) MustGetLongPollingUpdates() <-chan telego.Update {
	updates, err := c.UpdatesViaLongPolling(nil)
	if err != nil {
		log.Panicf("at MustGetLongPollingUpdates: %v", err)
	}
	return updates
}

// BotHandlerWrapper client for predicts.
type BotHandlerWrapper struct {
	*th.BotHandler
}

// MustNewBotHandler ...
func MustNewBotHandler(cw *ClientWrapper) *BotHandlerWrapper {
	if cw == nil {
		log.Panic("cw nil")
	}

	bh, err := th.NewBotHandler(cw.Bot, cw.MustGetLongPollingUpdates(), th.WithStopTimeout(time.Second))
	if err != nil {
		log.Panicf("at NewBotHandler: %v", err)
	}

	newCallbackCalendarHandler := calendar_wrapper.NewCallbackQueryForCalendarWrapper()

	bhw := &BotHandlerWrapper{
		BotHandler: bh,
	}

	bhw.RegisterHandlers(newCallbackCalendarHandler.CallbackQueryForCalendar)
	return bhw
}

// RegisterHandlers ...
func (bh *BotHandlerWrapper) RegisterHandlers(callbackCalendarHandler th.CallbackQueryHandler) {
	bh.HandleMessage(onStart, th.CommandEqual("start"))
	bh.HandleCallbackQuery(callbackCalendarHandler, th.AnyCallbackQueryWithMessage(), th.CallbackDataContains(constants.CallbackCalendar))
}

func onStart(bot *telego.Bot, message telego.Message) {
	if _, err := bot.SendMessage(tu.Message(
		tu.ID(utils.GetChatID(message)),
		fmt.Sprintf("Hello %s!\nLet's pretend an availability search has been done, and let's move on to choosing a day!", message.From.Username),
	).WithReplyMarkup(utils.GiveMainMenu(""))); err != nil {
		log.Printf("got err onStart send message: %v", err)
	}
}
