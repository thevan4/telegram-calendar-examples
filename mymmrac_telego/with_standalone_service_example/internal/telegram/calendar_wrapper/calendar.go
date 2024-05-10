package calendar_wrapper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/utils"
	pb "github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/pkg/telegram-calendar/telegram-calendar-examples/standalone_service" //nolint:lll // ok
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CallbackQueryForCalendarWrapper ...
type CallbackQueryForCalendarWrapper struct {
	globalCTX             context.Context
	calendarManagerClient pb.CalendarServiceClient
}

// MustNewClientForStandaloneService ...
func MustNewClientForStandaloneService(ctx context.Context, addr string) pb.CalendarServiceClient {
	if addr == "" {
		log.Fatal("at MustNewClientForStandaloneService: addr not set")
	}

	ctxWithDeadLine, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctxWithDeadLine,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cant grpc.DialContext for addr %v, error: %v", addr, err)
	}

	return pb.NewCalendarServiceClient(conn)
}

// MustNewCallbackQueryForCalendarWrapper ...
func MustNewCallbackQueryForCalendarWrapper(ctx context.Context, calendarStandaloneServiceAddr string) *CallbackQueryForCalendarWrapper {
	cmcl := MustNewClientForStandaloneService(ctx, calendarStandaloneServiceAddr)
	tn := time.Now()
	prevDay := tn.AddDate(0, 0, -1)

	_, err := cmcl.ApplyNewSettings(ctx,
		&pb.NewSettingsRequest{
			YearsBackForChoose: &pb.NewSettingsRequest_YearsBackForChoose{
				YearsBackForChoose: 0,
				ForceChoice:        true,
			},
			YearsForwardForChoose: &pb.NewSettingsRequest_YearsForwardForChoose{YearsForwardForChoose: 2},
			HomeButtonForBeauty:   &pb.NewSettingsRequest_HomeButtonForBeauty{HomeButtonForBeauty: "✈️"},
			PostfixForCurrentDay:  &pb.NewSettingsRequest_PostfixForCurrentDay{PostfixForCurrentDay: "🗓"},
			UnselectableDaysBeforeTime: &pb.NewSettingsRequest_UnselectableDaysBeforeTime{
				UnselectableDaysBeforeTime: prevDay.Format(time.RFC3339),
			},
			UnselectableDays: &pb.NewSettingsRequest_UnselectableDays{
				UnselectableDays: nil,
				ForceChoice:      true,
			},
		},
	)
	if err != nil {
		log.Fatalf("cant ApplyNewSettings for calendar manager, error: %v", err)
	}

	cw := &CallbackQueryForCalendarWrapper{calendarManagerClient: cmcl, globalCTX: ctx}

	go cw.AtNextMidnightChangeUnselectableDaysBeforeDate(tn)
	return cw
}

// AtNextMidnightChangeUnselectableDaysBeforeDate Logic to change the available days for selection so that the previous day is always unavailable for selection.
// This may still not work well if you don't take timezones correctly for your task.
func (cw *CallbackQueryForCalendarWrapper) AtNextMidnightChangeUnselectableDaysBeforeDate(currentDate time.Time) {
	nextMidnight := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day()+1, 0, 0, 0, 0, currentDate.Location())
	duration := nextMidnight.Sub(currentDate)

	go time.AfterFunc(duration, func() {
		prevDay := time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day()-1, 0, 0, 0, 0, nextMidnight.Location())
		if _, err := cw.calendarManagerClient.ApplyNewSettings(context.Background(), &pb.NewSettingsRequest{
			UnselectableDaysBeforeTime: &pb.NewSettingsRequest_UnselectableDaysBeforeTime{
				UnselectableDaysBeforeTime: prevDay.Format(time.RFC3339),
			},
		}); err != nil {
			// Retry logic looks good here.
			fmt.Printf("at AtNextMidnightChangeUnselectableDaysBeforeDate update UnselectableDaysBeforeDate fail: %v", err)
		}

		cw.AtNextMidnightChangeUnselectableDaysBeforeDate(nextMidnight)
	})
}

// CallbackQueryForCalendar ...
func (cw *CallbackQueryForCalendarWrapper) CallbackQueryForCalendar(bot *telego.Bot, query telego.CallbackQuery) {
	tn := time.Now()
	ctx, cancel := context.WithTimeout(cw.globalCTX, time.Second)
	defer cancel()

	generateCalendarKeyboardResponse, errGenerate := cw.calendarManagerClient.GenerateCalendar(ctx,
		&pb.GenerateCalendarRequest{
			CallbackPayload: query.Data,
			CurrentTime:     tn.Format(time.RFC3339),
		},
	)
	if errGenerate != nil {
		log.Printf("somehow generate calendar with grpc client at CallbackQueryForCalendar error: %v", errGenerate)
		return
	}

	// There may be additional processing logic
	if generateCalendarKeyboardResponse.GetIsUnselectableDay() {
		if err := bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).
			WithText("Day " + generateCalendarKeyboardResponse.GetSelectedDay().AsTime().Format("02.01.2006") + " is unselectable").WithShowAlert()); err != nil {
			log.Printf("at AnswerCallbackQuery for ShowAlert in UnselectableDay got error: %v", err)
		}
		return // show only alarm
	}

	// Place your day handler here!
	if generateCalendarKeyboardResponse.GetSelectedDay().IsValid() && !generateCalendarKeyboardResponse.GetSelectedDay().AsTime().IsZero() {
		responseMsg := &telego.EditMessageReplyMarkupParams{
			ChatID:    telego.ChatID{ID: utils.GetChatID(query)},
			MessageID: query.Message.MessageID,
			ReplyMarkup: utils.GiveMainMenu(fmt.Sprintf("Selected date %s. Back to calendar.",
				generateCalendarKeyboardResponse.GetSelectedDay().AsTime().Format("02.01.2006"))),
		}

		if _, err := bot.EditMessageReplyMarkup(responseMsg); err != nil {
			fmt.Printf("got err at bot.EditMessageReplyMarkup(responseMsg): %v", err)
		}
		return
	}

	if len(generateCalendarKeyboardResponse.GetInlineKeyboardMarkup().GetInlineKeyboard()) == 0 {
		// May silentDoNothingAction (aka "sdn"). For example, can occur when a blank cell without a date is selected.
		return
	}

	// The day was chosen and it was available for selection.
	replyKeyboard := protoInlineKeyboardMarkupToTelegoMsg(generateCalendarKeyboardResponse.GetInlineKeyboardMarkup())

	responseMsg := &telego.EditMessageReplyMarkupParams{
		ChatID:      telego.ChatID{ID: utils.GetChatID(query)},
		MessageID:   query.Message.MessageID,
		ReplyMarkup: &telego.InlineKeyboardMarkup{InlineKeyboard: replyKeyboard},
	}

	if _, err := bot.EditMessageReplyMarkup(responseMsg); err != nil {
		fmt.Printf("at CallbackQueryForCalendar at bot.EditMessageReplyMarkup %v", err)
	}
}

func protoInlineKeyboardMarkupToTelegoMsg(pbMarkup *pb.InlineKeyboardMarkup) [][]telego.InlineKeyboardButton {
	var result [][]telego.InlineKeyboardButton
	for _, row := range pbMarkup.GetInlineKeyboard() {
		var resultRow []telego.InlineKeyboardButton
		for _, pbButton := range row.GetButtons() {
			resultButton := telego.InlineKeyboardButton{
				Text:         pbButton.GetText(),
				CallbackData: pbButton.GetCallbackData(),
			}
			resultRow = append(resultRow, resultButton)
		}
		result = append(result, resultRow)
	}
	return result
}
