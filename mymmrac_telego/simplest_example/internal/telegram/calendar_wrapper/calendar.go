package calendar_wrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/utils"
	dbf "github.com/thevan4/telegram-calendar/day_button_former"
	"github.com/thevan4/telegram-calendar/generator"
	calendarManager "github.com/thevan4/telegram-calendar/manager"

	"github.com/mymmrac/telego"
)

// CallbackQueryForCalendarWrapper ...
type CallbackQueryForCalendarWrapper struct {
	calendarManager calendarManager.KeyboardManager
}

// NewCallbackQueryForCalendarWrapper ...
func NewCallbackQueryForCalendarWrapper() *CallbackQueryForCalendarWrapper {
	tn := time.Now()
	prevDay := tn.AddDate(0, 0, -1)
	cm := calendarManager.NewManager(
		generator.ChangeYearsForwardForChoose(2),
		generator.ChangeHomeButtonForBeauty("✈️"),
		generator.NewButtonsTextWrapper(
			// 0 current day is unavailable, -1 since last day
			dbf.ChangeUnselectableDaysBeforeDate(prevDay),
			dbf.ChangePostfixForCurrentDay("🗓"),
		),
	)

	cw := &CallbackQueryForCalendarWrapper{calendarManager: cm}
	go cw.AtNextMidnightChangeUnselectableDaysBeforeDate(tn)

	return cw
}

// AtNextMidnightChangeUnselectableDaysBeforeDate Logic to change the available days for selection so that the previous day is always unavailable for selection.
// This may still not work well if you don't take timezones correctly for your task.
func (cw *CallbackQueryForCalendarWrapper) AtNextMidnightChangeUnselectableDaysBeforeDate(currentDate time.Time) {
	nextMidnight := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day()+1, 0, 0, 0, 0, currentDate.Location())
	duration := nextMidnight.Sub(currentDate)
	log.Printf("currentDate %v, nextMidnight %v, duration %v", currentDate, nextMidnight, duration)
	go time.AfterFunc(duration, func() {
		prevDay := time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day()-1, 0, 0, 0, 0, nextMidnight.Location())
		cw.calendarManager.ApplyNewOptions(generator.ApplyNewOptionsForButtonsTextWrapper(
			dbf.ChangeUnselectableDaysBeforeDate(prevDay),
		))

		cw.AtNextMidnightChangeUnselectableDaysBeforeDate(nextMidnight)
	})
}

// CallbackQueryForCalendar ...
func (cw *CallbackQueryForCalendarWrapper) CallbackQueryForCalendar(bot *telego.Bot, query telego.CallbackQuery) {
	// For real use, it is better to throw the necessary timezone (your local one, take from the user from the database, etc.)
	now := time.Now()
	tn := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	generateCalendarKeyboardResponse := cw.calendarManager.GenerateCalendarKeyboard(query.Data, tn)

	// There may be additional processing logic
	if generateCalendarKeyboardResponse.IsUnselectableDay {
		if err := bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).
			WithText("Day " + generateCalendarKeyboardResponse.SelectedDay.Format("02.01.2006") + " is unselectable").WithShowAlert()); err != nil {
			log.Printf("at AnswerCallbackQuery for ShowAlert in UnselectableDay got error: %v", err)
		}
		return // show only alarm
	}

	// Place your day handler here!
	if !generateCalendarKeyboardResponse.SelectedDay.IsZero() {
		responseMsg := &telego.EditMessageReplyMarkupParams{
			ChatID:    telego.ChatID{ID: utils.GetChatID(query)},
			MessageID: query.Message.MessageID,
			ReplyMarkup: utils.GiveMainMenu(fmt.Sprintf("Selected date %s. Back to calendar.",
				generateCalendarKeyboardResponse.SelectedDay.Format("02.01.2006"))),
		}

		if _, err := bot.EditMessageReplyMarkup(responseMsg); err != nil {
			fmt.Printf("got err at bot.EditMessageReplyMarkup(responseMsg): %v", err)
		}
		return
	}

	if len(generateCalendarKeyboardResponse.InlineKeyboardMarkup.InlineKeyboard) == 0 {
		// May silentDoNothingAction (aka "sdn"). For example, can occur when a blank cell without a date is selected.
		return
	}

	// The day was chosen and it was available for selection.
	b, err := json.Marshal(generateCalendarKeyboardResponse.InlineKeyboardMarkup)
	if err != nil {
		log.Fatalf("at CallbackQueryForCalendar json.Marshal(resp) error: %v", err)
	}

	replyKeyboard := new(telego.InlineKeyboardMarkup)
	if err = json.Unmarshal(b, replyKeyboard); err != nil {
		log.Fatalf("at CallbackQueryForCalendar json.Unmarshal error: %v", err)
	}

	responseMsg := &telego.EditMessageReplyMarkupParams{
		ChatID:      telego.ChatID{ID: utils.GetChatID(query)},
		MessageID:   query.Message.MessageID,
		ReplyMarkup: replyKeyboard,
	}

	_, err = bot.EditMessageReplyMarkup(responseMsg)
	if err != nil {
		fmt.Printf("at CallbackQueryForCalendar at bot.EditMessageReplyMarkup %v", err)
	}
}
