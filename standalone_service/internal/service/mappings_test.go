package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	telegramCalendarPb "github.com/thevan4/telegram-calendar-examples/standalone_service/pkg/telegram-calendar"
	calendarManager "github.com/thevan4/telegram-calendar/manager"
	"github.com/thevan4/telegram-calendar/models"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapGenerateCalendarKeyboardToResponse(t *testing.T) {
	t.Parallel()
	type args struct {
		inlineKeyboard models.InlineKeyboardMarkup
		selectedDay    time.Time
	}
	tests := []struct {
		name             string
		args             args
		wantText         string
		wantCallbackData string
		wantSelectedDay  *timestamppb.Timestamp
		wantErr          bool
	}{
		{
			name: "valid data",
			args: args{
				inlineKeyboard: models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{Text: "button1", CallbackData: "data1"},
						},
					},
				},
				selectedDay: time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC),
			},
			wantText:         "button1",
			wantCallbackData: "data1",
			wantSelectedDay:  timestamppb.New(time.Date(2022, 10, 5, 0, 0, 0, 0, time.UTC)),
			wantErr:          false,
		},
		{
			name: "zero day",
			args: args{
				inlineKeyboard: models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{Text: "button1", CallbackData: "data1"},
						},
					},
				},
				selectedDay: time.Time{},
			},
			wantText:         "button1",
			wantCallbackData: "data1",
			wantSelectedDay:  nil,
			wantErr:          false,
		},
	}

	for _, tmpTT := range tests {
		tt := tmpTT
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			response := mapGenerateCalendarKeyboardToResponse(tt.args.inlineKeyboard, tt.args.selectedDay)

			assert.NotNil(t, response)
			assert.Equal(t, tt.wantText, response.InlineKeyboardMarkup.InlineKeyboard[0].Buttons[0].Text)
			assert.Equal(t, tt.wantCallbackData, response.InlineKeyboardMarkup.InlineKeyboard[0].Buttons[0].CallbackData)
			if tt.wantSelectedDay != nil {
				assert.NotNil(t, response.SelectedDay)
				assert.Equal(t, tt.wantSelectedDay, response.SelectedDay, "selected date does not match the expected date")
			} else {
				assert.Nil(t, response.SelectedDay, "the selected date must be nil when the date is zero")
			}
		})
	}
}

func TestMapCurrentConfigToResponse(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	type args struct {
		currentConfig calendarManager.FlatConfig
	}
	tests := []struct {
		name string
		args args
		want *telegramCalendarPb.GetSettingsResponse
	}{
		{
			name: "test with full data",
			args: args{
				currentConfig: calendarManager.FlatConfig{
					YearsBackForChoose:    3,
					YearsForwardForChoose: 5,
					SumYearsForChoose:     8,
					DaysNames:             [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
					MonthNames: [12]string{"Jan1", "Feb2", "Mar3", "Apr4", "May5", "Jun6", "Jul7", "Aug8",
						"Sep9", "Oct10", "Nov11", "Dec12"},
					HomeButtonForBeauty:        "adba",
					PrefixForCurrentDay:        "C:",
					PostfixForCurrentDay:       ":C",
					PrefixForNonSelectedDay:    "N:",
					PostfixForNonSelectedDay:   ":N",
					PrefixForPickDay:           "P:",
					PostfixForPickDay:          ":P",
					UnselectableDaysBeforeTime: now,
					UnselectableDaysAfterTime:  now.Add(24 * time.Hour),
					UnselectableDays:           map[time.Time]struct{}{now: {}},
				},
			},
			want: &telegramCalendarPb.GetSettingsResponse{
				YearsBackForChoose:    3,
				YearsForwardForChoose: 5,
				SumYearsForChoose:     8,
				DaysNames:             []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
				MonthNames: []string{"Jan1", "Feb2", "Mar3", "Apr4", "May5", "Jun6", "Jul7", "Aug8",
					"Sep9", "Oct10", "Nov11", "Dec12"},
				HomeButtonForBeauty:        "adba",
				PrefixForCurrentDay:        "C:",
				PostfixForCurrentDay:       ":C",
				PrefixForNonSelectedDay:    "N:",
				PostfixForNonSelectedDay:   ":N",
				PrefixForPickDay:           "P:",
				PostfixForPickDay:          ":P",
				UnselectableDaysBeforeTime: timestamppb.New(now),
				UnselectableDaysAfterTime:  timestamppb.New(now.Add(24 * time.Hour)),
				UnselectableDays:           convertTimeMapToProtoMap(map[time.Time]struct{}{now: {}}),
			},
		},
	}

	for _, tmpTT := range tests {
		tt := tmpTT
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapCurrentConfigToResponse(tt.args.currentConfig)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertTimeMapToProtoMap(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	type args struct {
		tm map[time.Time]struct{}
	}
	tests := []struct {
		name string
		args args
		want map[string]*emptypb.Empty
	}{
		{
			name: "empty map",
			args: args{
				tm: map[time.Time]struct{}{},
			},
			want: map[string]*emptypb.Empty{},
		},
		{
			name: "single element",
			args: args{
				tm: map[time.Time]struct{}{
					now: {},
				},
			},
			want: map[string]*emptypb.Empty{
				now.Format(time.RFC3339): {},
			},
		},
		{
			name: "multiple elements",
			args: args{
				tm: map[time.Time]struct{}{
					earlier: {},
					now:     {},
					later:   {},
				},
			},
			want: map[string]*emptypb.Empty{
				earlier.Format(time.RFC3339): {},
				now.Format(time.RFC3339):     {},
				later.Format(time.RFC3339):   {},
			},
		},
	}

	for _, tmpTT := range tests {
		tt := tmpTT
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := convertTimeMapToProtoMap(tt.args.tm)
			assert.Equal(t, tt.want, got, "convertTimeMapToProtoMap() did not return the expected map")
		})
	}
}

func TestMapToConfigCallbacks(t *testing.T) {
	type args struct {
		req *telegramCalendarPb.NewSettingsRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantLen int
	}{
		{
			name: "nil request",
			args: args{
				req: nil,
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "empty request",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{},
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "Unix epoch time with ForceChoice true",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					UnselectableDaysBeforeTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysBeforeTime{
						UnselectableDaysBeforeTime: &timestamppb.Timestamp{Seconds: unixEpochTime.Unix()},
						ForceChoice:                true,
					},
					UnselectableDaysAfterTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysAfterTime{
						UnselectableDaysAfterTime: &timestamppb.Timestamp{Seconds: unixEpochTime.Unix()},
						ForceChoice:               false,
					},
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "Protobuf default time with ForceChoice true",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					UnselectableDaysBeforeTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysBeforeTime{
						UnselectableDaysBeforeTime: &timestamppb.Timestamp{Seconds: unixEpochTime.Unix()},
						ForceChoice:                false,
					},
					UnselectableDaysAfterTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysAfterTime{
						UnselectableDaysAfterTime: &timestamppb.Timestamp{Seconds: unixEpochTime.Unix()},
						ForceChoice:               true,
					},
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "valid request with multiple callbacks",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
						ForceChoice:        true,
						YearsBackForChoose: 3,
					},
					YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
						ForceChoice:           true,
						YearsForwardForChoose: 5,
					},
					DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
						ForceChoice: true,
						DayNames: []*telegramCalendarPb.NewSettingsRequest_DayName{
							{DayName: "Mon"}, {DayName: "Tue"}, {DayName: "Wed"}, {DayName: "Thu"},
							{DayName: "Fri"}, {DayName: "Sat"}, {DayName: "Sun"},
						},
					},
					MonthNames: &telegramCalendarPb.NewSettingsRequest_MonthNames{
						ForceChoice: true,
						MonthNames: []*telegramCalendarPb.NewSettingsRequest_MonthName{
							{MonthName: "Jan"}, {MonthName: "Feb"}, {MonthName: "Mar"}, {MonthName: "Apr"},
							{MonthName: "May"}, {MonthName: "Jun"}, {MonthName: "Jul"}, {MonthName: "Aug"},
							{MonthName: "Sep"}, {MonthName: "Oct"}, {MonthName: "Nov"}, {MonthName: "Dec"},
						},
					},
				},
			},
			wantErr: false,
			wantLen: 4, // 2 for years, 1 for days, 1 for months
		},
		{
			name: "complex test: all callbacks",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
						ForceChoice:        false,
						YearsBackForChoose: 3,
					},
					YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
						ForceChoice:           false,
						YearsForwardForChoose: 5,
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
							{MonthName: "Jan"}, {MonthName: "Feb"}, {MonthName: "Mar"}, {MonthName: "Apr"},
							{MonthName: "May"}, {MonthName: "Jun"}, {MonthName: "Jul"}, {MonthName: "Aug"},
							{MonthName: "Sep"}, {MonthName: "Oct"}, {MonthName: "Nov"}, {MonthName: "Dec"},
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
				},
			},
			wantErr: false,
			wantLen: 14,
		},
		{
			name: "complex test: all with zero values and force set",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
						ForceChoice: true,
					},
					YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
						ForceChoice: true,
					},
					DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
						ForceChoice: true,
					},
					MonthNames: &telegramCalendarPb.NewSettingsRequest_MonthNames{
						ForceChoice: true,
					},
					HomeButtonForBeauty: &telegramCalendarPb.NewSettingsRequest_HomeButtonForBeauty{
						ForceChoice: true,
					},
					PrefixForCurrentDay: &telegramCalendarPb.NewSettingsRequest_PrefixForCurrentDay{
						ForceChoice: true,
					},
					PostfixForCurrentDay: &telegramCalendarPb.NewSettingsRequest_PostfixForCurrentDay{
						ForceChoice: true,
					},
					PrefixForNonSelectedDay: &telegramCalendarPb.NewSettingsRequest_PrefixForNonSelectedDay{
						ForceChoice: true,
					},
					PostfixForNonSelectedDay: &telegramCalendarPb.NewSettingsRequest_PostfixForNonSelectedDay{
						ForceChoice: true,
					},
					PrefixForPickDay: &telegramCalendarPb.NewSettingsRequest_PrefixForPickDay{
						ForceChoice: true,
					},
					PostfixForPickDay: &telegramCalendarPb.NewSettingsRequest_PostfixForPickDay{
						ForceChoice: true,
					},
					UnselectableDaysBeforeTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysBeforeTime{
						ForceChoice: true,
					},
					UnselectableDaysAfterTime: &telegramCalendarPb.NewSettingsRequest_UnselectableDaysAfterTime{
						ForceChoice: true,
					},
					UnselectableDays: &telegramCalendarPb.NewSettingsRequest_UnselectableDays{
						ForceChoice: true,
					},
				},
			},
			wantErr: false,
			wantLen: 12, // not 14, because at days names and months names slices must exist also for zero values
		},
		{
			name: "set some days and mounts names force empty",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
						ForceChoice: true,
						DayNames: []*telegramCalendarPb.NewSettingsRequest_DayName{
							{DayName: ""}, {DayName: "Tue"}, {DayName: "Wed"}, {DayName: "Thu"},
							{DayName: "Fri"}, {DayName: ""}, {DayName: ""},
						},
					},
					MonthNames: &telegramCalendarPb.NewSettingsRequest_MonthNames{
						ForceChoice: true,
						MonthNames: []*telegramCalendarPb.NewSettingsRequest_MonthName{
							{MonthName: "Jan"}, {MonthName: ""}, {MonthName: "Mar"}, {MonthName: ""},
							{MonthName: "May"}, {MonthName: ""}, {MonthName: ""}, {MonthName: "Aug"},
							{MonthName: ""}, {MonthName: ""}, {MonthName: "Nov"}, {MonthName: "Dec"},
						},
					},
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "invalid request with to many days",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
						ForceChoice:        true,
						YearsBackForChoose: 3,
					},
					YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
						ForceChoice:           true,
						YearsForwardForChoose: 5,
					},
					DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
						ForceChoice: true,
						DayNames: []*telegramCalendarPb.NewSettingsRequest_DayName{
							{DayName: "Mon"}, {DayName: "Tue"}, {DayName: "Wed"}, {DayName: "Thu"},
							{DayName: "Fri"}, {DayName: "Sat"}, {DayName: "Sun"}, {DayName: "Wth"},
						},
					},
				},
			},
			wantErr: true,
			wantLen: 2, // 2 for years, 1 fail (error) for days
		},
		{
			name: "invalid request with to many days",
			args: args{
				req: &telegramCalendarPb.NewSettingsRequest{
					YearsBackForChoose: &telegramCalendarPb.NewSettingsRequest_YearsBackForChoose{
						ForceChoice:        true,
						YearsBackForChoose: 3,
					},
					YearsForwardForChoose: &telegramCalendarPb.NewSettingsRequest_YearsForwardForChoose{
						ForceChoice:           true,
						YearsForwardForChoose: 5,
					},
					DayNames: &telegramCalendarPb.NewSettingsRequest_DayNames{
						ForceChoice: true,
						DayNames: []*telegramCalendarPb.NewSettingsRequest_DayName{
							{DayName: "Mon"}, {DayName: "Tue"}, {DayName: "Wed"}, {DayName: "Thu"},
							{DayName: "Fri"}, {DayName: "Sat"}, {DayName: "Sun"},
						},
					},
					MonthNames: &telegramCalendarPb.NewSettingsRequest_MonthNames{
						ForceChoice: true,
						MonthNames: []*telegramCalendarPb.NewSettingsRequest_MonthName{
							{MonthName: "Jan"}, {MonthName: "Feb"}, {MonthName: "Mar"}, {MonthName: "Apr"},
							{MonthName: "May"}, {MonthName: "Jun"}, {MonthName: "Jul"}, {MonthName: "Aug"},
							{MonthName: "Sep"}, {MonthName: "Oct"}, {MonthName: "Nov"}, {MonthName: "Dec"},
							{MonthName: "Ouch"},
						},
					},
				},
			},
			wantErr: true,
			wantLen: 3, // 2 for years, 1 for days,  1 fail (error) for months
		},
	}

	for _, tmpTT := range tests {
		tt := tmpTT
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mapToConfigCallbacks(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapToConfigCallbacks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Errorf("mapToConfigCallbacks() got len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
