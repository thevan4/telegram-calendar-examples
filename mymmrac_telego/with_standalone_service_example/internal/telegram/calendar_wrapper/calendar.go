package calendar_wrapper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/with_standalone_service_example/internal/telegram/utils"
	pb "github.com/thevan4/telegram-calendar-examples/mymmrac_telego/with_standalone_service_example/pkg/telegram-calendar/telegram-calendar-examples/standalone_service" //nolint:lll // ok
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CallbackQueryForCalendarWrapper ...
type CallbackQueryForCalendarWrapper struct {
	globalCTX             context.Context
	conn                  *grpc.ClientConn
	calendarManagerClient pb.CalendarServiceClient
	cncgrpc               context.CancelFunc
}

// Close closes the underlying gRPC connection.
func (cw *CallbackQueryForCalendarWrapper) Close() error {
	defer cw.cncgrpc()

	if cw.conn != nil {
		return cw.conn.Close()
	}
	return nil
}

// MustNewClientForStandaloneService ...
func MustNewClientForStandaloneService(ctx context.Context, addr string) (*grpc.ClientConn, pb.CalendarServiceClient, context.CancelFunc) {
	if addr == "" {
		log.Fatal("at MustNewClientForStandaloneService: addr not set")
	}

	ctxWithDeadLine, cancel := context.WithTimeout(ctx, time.Second)

	conn, err := grpc.DialContext(
		ctxWithDeadLine,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cant grpc.NewClient for addr %v, error: %v", addr, err)
	}

	return conn, pb.NewCalendarServiceClient(conn), cancel
}

// MustNewCallbackQueryForCalendarWrapper ...
func MustNewCallbackQueryForCalendarWrapper(ctx context.Context, calendarStandaloneServiceAddr string) *CallbackQueryForCalendarWrapper {
	conn, cmcl, cncgrpc := MustNewClientForStandaloneService(ctx, calendarStandaloneServiceAddr)
	tzName := os.Getenv("TIMEZONE")
	if tzName == "" {
		tzName = "Etc/GMT-3"
	}
	tz, errParseTimezone := time.LoadLocation(tzName)
	if errParseTimezone != nil {
		log.Fatalf("parse new timezone error: %v", errParseTimezone)
	}
	tn := time.Now().In(tz)
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
			Timezone: &pb.NewSettingsRequest_Timezone{Timezone: tn.Location().String()},
		},
	)
	if err != nil {
		log.Fatalf("cant ApplyNewSettings for calendar manager, error: %v", err)
	}

	cw := &CallbackQueryForCalendarWrapper{calendarManagerClient: cmcl, globalCTX: ctx, conn: conn, cncgrpc: cncgrpc}

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
		if _, err := cw.calendarManagerClient.ApplyNewSettings(cw.globalCTX, &pb.NewSettingsRequest{
			UnselectableDaysBeforeTime: &pb.NewSettingsRequest_UnselectableDaysBeforeTime{
				UnselectableDaysBeforeTime: prevDay.Format(time.RFC3339),
			},
		}); err != nil {
			// Retry logic looks good here.
			log.Printf("at AtNextMidnightChangeUnselectableDaysBeforeDate update UnselectableDaysBeforeDate fail: %v\n", err)
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
			log.Printf("got err at bot.EditMessageReplyMarkup(responseMsg): %v\n", err)
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
		log.Printf("at CallbackQueryForCalendar at bot.EditMessageReplyMarkup %v\n", err)
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
