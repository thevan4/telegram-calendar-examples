// It's actually almost integration tests.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	telegramCalendarPb "github.com/thevan4/telegram-calendar-examples/standalone_service/pkg/telegram-calendar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, s string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}
}

func TestGrpcServerAndClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := grpc.NewServer()
	defer s.Stop()
	telegramCalendarPb.RegisterCalendarServiceServer(s, NewTelegramCalendarGRPCService())

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("server exited with error: %v", err)
		}
	}()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer(lis)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer func() {
		if errCloseConn := conn.Close(); errCloseConn != nil {
			t.Fatalf("somehow conn.Close error: %v", errCloseConn)
		}
	}()

	client := telegramCalendarPb.NewCalendarServiceClient(conn)

	// can't be run separately because there's a configuration conflict. Don't want to solve the problem with mutex.
	testGenerateCalendar(ctx, t, client)
	fmt.Println("WRF")
	// Read default config
	defaultConfig, errGetSettings := client.GetSettings(ctx, new(emptypb.Empty))
	if errGetSettings != nil {
		t.Fatalf("GetSettings (default) failed: %v", errGetSettings)
	}

	// and compare with expectedDefaultConfig
	expectedDefaultConfig := &telegramCalendarPb.GetSettingsResponse{
		YearsBackForChoose:         0,
		YearsForwardForChoose:      3,
		SumYearsForChoose:          3,
		DaysNames:                  []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"},
		MonthNames:                 []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		HomeButtonForBeauty:        "🏩",
		PrefixForCurrentDay:        "",
		PostfixForCurrentDay:       "",
		PrefixForNonSelectedDay:    "",
		PostfixForNonSelectedDay:   "❌",
		PrefixForPickDay:           "",
		PostfixForPickDay:          "",
		UnselectableDaysBeforeTime: timestamppb.New(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
		UnselectableDaysAfterTime:  timestamppb.New(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)),
		UnselectableDays:           map[string]*emptypb.Empty{},
	}

	// not protojson.Marshal because https://github.com/golang/protobuf/issues/1351
	gotDefaultJson, errGotDefaultJsonMarshal := json.Marshal(defaultConfig)
	if errGotDefaultJsonMarshal != nil {
		t.Fatalf("failed to marshal got settings: %v", errGotDefaultJsonMarshal)
	}
	wantDefaultJson, errWantDefaultJsonMarshal := json.Marshal(expectedDefaultConfig)
	if errWantDefaultJsonMarshal != nil {
		t.Fatalf("failed to marshal expected settings: %v", errWantDefaultJsonMarshal)
	}

	if string(gotDefaultJson) != string(wantDefaultJson) {
		t.Fatalf("settings mismatch: got %s, want %s", gotDefaultJson, wantDefaultJson)
	}

	// apply new settings
	requestApplyNewSettings := &telegramCalendarPb.NewSettingsRequest{
		YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
			ForceChoice:        false,
			YearsBackForChoose: 1,
		},
		YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
			ForceChoice:           false,
			YearsForwardForChoose: 2,
		},
		DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
			ForceChoice: false,
			DayNames: []*telegramCalendarPb.NewSettingsRequest_DayName{
				{DayName: "Mon"}, {DayName: "Tue"}, {DayName: "Wed"}, {DayName: "Thu"},
				{DayName: "Fri"}, {DayName: "Sat"}, {DayName: "Sun"},
			},
		},
		MonthNames: &telegramCalendarPb.NewSettingsRequest_MonthNames{
			ForceChoice: false,
			MonthNames: []*telegramCalendarPb.NewSettingsRequest_MonthName{
				{MonthName: "Jan1"}, {MonthName: "Feb2"}, {MonthName: "Mar3"},
				{MonthName: "Apr4"}, {MonthName: "May5"}, {MonthName: "Jun6"},
				{MonthName: "Jul7"}, {MonthName: "Aug8"}, {MonthName: "Sep9"},
				{MonthName: "Oct10"}, {MonthName: "Nov11"}, {MonthName: "Dec12"},
			},
		},
		HomeButtonForBeauty: &telegramCalendarPb.NewSettingsRequest_HomeButtonForBeauty{
			ForceChoice:         false,
			HomeButtonForBeauty: "poop",
		},
		PrefixForCurrentDay: &telegramCalendarPb.NewSettingsRequest_PrefixForCurrentDay{
			ForceChoice:         false,
			PrefixForCurrentDay: "P:",
		},
		PostfixForCurrentDay: &telegramCalendarPb.NewSettingsRequest_PostfixForCurrentDay{
			ForceChoice:          false,
			PostfixForCurrentDay: ":P",
		},
		PrefixForNonSelectedDay: &telegramCalendarPb.NewSettingsRequest_PrefixForNonSelectedDay{
			ForceChoice:             false,
			PrefixForNonSelectedDay: "N:",
		},
		PostfixForNonSelectedDay: &telegramCalendarPb.NewSettingsRequest_PostfixForNonSelectedDay{
			ForceChoice:              false,
			PostfixForNonSelectedDay: ":N",
		},
		PrefixForPickDay: &telegramCalendarPb.NewSettingsRequest_PrefixForPickDay{
			ForceChoice:      false,
			PrefixForPickDay: "P:",
		},
		PostfixForPickDay: &telegramCalendarPb.NewSettingsRequest_PostfixForPickDay{
			ForceChoice:       false,
			PostfixForPickDay: ":P",
		},
		UnselectableDaysBeforeTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysBeforeTime{
			ForceChoice:                false,
			UnselectableDaysBeforeTime: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		UnselectableDaysAfterTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysAfterTime{
			ForceChoice:               false,
			UnselectableDaysAfterTime: timestamppb.New(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		UnselectableDays: &telegramCalendarPb.NewSettingsRequest_UnselectableDays{
			ForceChoice: false,
			UnselectableDays: map[string]*emptypb.Empty{
				`2024-01-02T00:00:00.00Z`: {}},
		},
	}

	_, errApplyNewSettings := client.ApplyNewSettings(ctx, requestApplyNewSettings)
	if errApplyNewSettings != nil {
		t.Fatalf("ApplyNewSettings failed: %v", errApplyNewSettings)
	}

	// and compare response
	customConfig, errGetSettingsCustom := client.GetSettings(ctx, new(emptypb.Empty))
	if errGetSettingsCustom != nil {
		t.Fatalf("GetSettings (custom) failed: %v", errGetSettingsCustom)
	}

	expectedCustomConfig := &telegramCalendarPb.GetSettingsResponse{
		YearsBackForChoose:    1,
		YearsForwardForChoose: 2,
		SumYearsForChoose:     3,
		DaysNames:             []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		MonthNames: []string{"Jan1", "Feb2", "Mar3", "Apr4", "May5", "Jun6", "Jul7", "Aug8",
			"Sep9", "Oct10", "Nov11", "Dec12"},
		HomeButtonForBeauty:        "poop",
		PrefixForCurrentDay:        "P:",
		PostfixForCurrentDay:       ":P",
		PrefixForNonSelectedDay:    "N:",
		PostfixForNonSelectedDay:   ":N",
		PrefixForPickDay:           "P:",
		PostfixForPickDay:          ":P",
		UnselectableDaysBeforeTime: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		UnselectableDaysAfterTime:  timestamppb.New(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)),
		UnselectableDays:           map[string]*emptypb.Empty{`2024-01-02T00:00:00Z`: {}},
	}

	gotCustomJson, errGotCustomJsonMarshal := json.Marshal(customConfig)
	if errGotCustomJsonMarshal != nil {
		t.Fatalf("failed to marshal got settings: %v", errGotCustomJsonMarshal)
	}
	wantCustomJson, errWantCustomJsonMarshal := json.Marshal(expectedCustomConfig)
	if errWantCustomJsonMarshal != nil {
		t.Fatalf("failed to marshal expected settings: %v", errWantCustomJsonMarshal)
	}

	if string(gotCustomJson) != string(wantCustomJson) {
		t.Fatalf("settings mismatch: got %s, want %s", gotCustomJson, wantCustomJson)
	}
}

func testGenerateCalendar(ctx context.Context, t *testing.T, client telegramCalendarPb.CalendarServiceClient) {
	curTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	firstGenResponse, errFirstGenResponse := client.GenerateCalendar(ctx, &telegramCalendarPb.GenerateCalendarRequest{
		CallbackPayload: "",
		CurrentTime:     timestamppb.New(curTime),
	})
	if errFirstGenResponse != nil {
		t.Fatalf("GenerateCalendar witn empty CallbackPayload and mock currents time %v error: %v", curTime, errFirstGenResponse)
	}

	// not protojson.Marshal because https://github.com/golang/protobuf/issues/1351
	gotFirstGenResponseRaw, errGotFirstGenResponseMarshal := json.Marshal(firstGenResponse)
	if errGotFirstGenResponseMarshal != nil {
		t.Fatalf("failed to marshal got settings: %v", errGotFirstGenResponseMarshal)
	}

	expectedFirstGenResponse := `{"inline_keyboard_markup":{"inline_keyboard":[{"buttons":[{"text":"«","callback_data":"calendar/pry_00.01.2025"},{"text":"\u003c","callback_data":"calendar/prm_00.01.2025"},{"text":"Jan","callback_data":"calendar/sem_00.01.2025"},{"text":"🏩","callback_data":"calendar/sdn_00.01.2025"},{"text":"2025","callback_data":"calendar/sey_00.01.2025"},{"text":"\u003e","callback_data":"calendar/nem_00.01.2025"},{"text":"»","callback_data":"calendar/ney_00.01.2025"}]},{"buttons":[{"text":"Mo","callback_data":"calendar/sdn_00.01.2025"},{"text":"Tu","callback_data":"calendar/sdn_00.01.2025"},{"text":"We","callback_data":"calendar/sdn_00.01.2025"},{"text":"Th","callback_data":"calendar/sdn_00.01.2025"},{"text":"Fr","callback_data":"calendar/sdn_00.01.2025"},{"text":"Sa","callback_data":"calendar/sdn_00.01.2025"},{"text":"Su","callback_data":"calendar/sdn_00.01.2025"}]},{"buttons":[{"text":" ","callback_data":"calendar/sdn_00.01.2025"},{"text":" ","callback_data":"calendar/sdn_00.01.2025"},{"text":"1","callback_data":"calendar/sed_01.01.2025"},{"text":"2","callback_data":"calendar/sed_02.01.2025"},{"text":"3","callback_data":"calendar/sed_03.01.2025"},{"text":"4","callback_data":"calendar/sed_04.01.2025"},{"text":"5","callback_data":"calendar/sed_05.01.2025"}]},{"buttons":[{"text":"6","callback_data":"calendar/sed_06.01.2025"},{"text":"7","callback_data":"calendar/sed_07.01.2025"},{"text":"8","callback_data":"calendar/sed_08.01.2025"},{"text":"9","callback_data":"calendar/sed_09.01.2025"},{"text":"10","callback_data":"calendar/sed_10.01.2025"},{"text":"11","callback_data":"calendar/sed_11.01.2025"},{"text":"12","callback_data":"calendar/sed_12.01.2025"}]},{"buttons":[{"text":"13","callback_data":"calendar/sed_13.01.2025"},{"text":"14","callback_data":"calendar/sed_14.01.2025"},{"text":"15","callback_data":"calendar/sed_15.01.2025"},{"text":"16","callback_data":"calendar/sed_16.01.2025"},{"text":"17","callback_data":"calendar/sed_17.01.2025"},{"text":"18","callback_data":"calendar/sed_18.01.2025"},{"text":"19","callback_data":"calendar/sed_19.01.2025"}]},{"buttons":[{"text":"20","callback_data":"calendar/sed_20.01.2025"},{"text":"21","callback_data":"calendar/sed_21.01.2025"},{"text":"22","callback_data":"calendar/sed_22.01.2025"},{"text":"23","callback_data":"calendar/sed_23.01.2025"},{"text":"24","callback_data":"calendar/sed_24.01.2025"},{"text":"25","callback_data":"calendar/sed_25.01.2025"},{"text":"26","callback_data":"calendar/sed_26.01.2025"}]},{"buttons":[{"text":"27","callback_data":"calendar/sed_27.01.2025"},{"text":"28","callback_data":"calendar/sed_28.01.2025"},{"text":"29","callback_data":"calendar/sed_29.01.2025"},{"text":"30","callback_data":"calendar/sed_30.01.2025"},{"text":"31","callback_data":"calendar/sed_31.01.2025"},{"text":" ","callback_data":"calendar/sdn_00.01.2025"},{"text":" ","callback_data":"calendar/sdn_00.01.2025"}]}]}}`
	if string(gotFirstGenResponseRaw) != expectedFirstGenResponse {
		t.Fatalf("settings mismatch: got %s, want %v", gotFirstGenResponseRaw, expectedFirstGenResponse)
	}
}
