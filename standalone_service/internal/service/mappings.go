package service

import (
	"fmt"
	"time"

	telegramCalendarPb "github.com/thevan4/telegram-calendar-examples/standalone_service/pkg/telegram-calendar"
	telegramDayButtonFormer "github.com/thevan4/telegram-calendar/day_button_former"
	calendarGenerator "github.com/thevan4/telegram-calendar/generator"
	calendarManager "github.com/thevan4/telegram-calendar/manager"
	"github.com/thevan4/telegram-calendar/models"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	defaultProtoTime = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	unixEpochTime    = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
)

func mapGenerateCalendarKeyboardToResponse(
	inlineKeyboardMarkup models.InlineKeyboardMarkup,
	selectedDay time.Time,
) *telegramCalendarPb.GenerateCalendarResponse {
	// mapping InlineKeyboardMarkup
	inlineKeyboardMarkupResponse := &telegramCalendarPb.InlineKeyboardMarkup{
		InlineKeyboard: make([]*telegramCalendarPb.InlineKeyboardMarkup_InlineKeyboardRow, len(inlineKeyboardMarkup.InlineKeyboard)),
	}
	for i, row := range inlineKeyboardMarkup.InlineKeyboard {
		convertedRow := &telegramCalendarPb.InlineKeyboardMarkup_InlineKeyboardRow{
			Buttons: make([]*telegramCalendarPb.InlineKeyboardMarkup_InlineKeyboardButton, len(row)),
		}
		for j, button := range row {
			convertedButton := &telegramCalendarPb.InlineKeyboardMarkup_InlineKeyboardButton{
				Text:         button.Text,
				CallbackData: button.CallbackData,
			}
			convertedRow.Buttons[j] = convertedButton
		}
		inlineKeyboardMarkupResponse.InlineKeyboard[i] = convertedRow
	}

	// mapping result
	result := &telegramCalendarPb.GenerateCalendarResponse{InlineKeyboardMarkup: inlineKeyboardMarkupResponse}

	if !selectedDay.IsZero() {
		result.SelectedDay = timestamppb.New(selectedDay)
	}

	return result
}

func mapCurrentConfigToResponse(currentConfig calendarManager.FlatConfig) *telegramCalendarPb.GetSettingsResponse {
	result := &telegramCalendarPb.GetSettingsResponse{
		YearsBackForChoose:         int64(currentConfig.YearsBackForChoose),
		YearsForwardForChoose:      int64(currentConfig.YearsForwardForChoose),
		SumYearsForChoose:          int64(currentConfig.SumYearsForChoose),
		DaysNames:                  currentConfig.DaysNames[:],
		MonthNames:                 currentConfig.MonthNames[:],
		HomeButtonForBeauty:        currentConfig.HomeButtonForBeauty,
		PrefixForCurrentDay:        currentConfig.PrefixForCurrentDay,
		PostfixForCurrentDay:       currentConfig.PostfixForCurrentDay,
		PrefixForNonSelectedDay:    currentConfig.PrefixForNonSelectedDay,
		PostfixForNonSelectedDay:   currentConfig.PostfixForNonSelectedDay,
		PrefixForPickDay:           currentConfig.PrefixForPickDay,
		PostfixForPickDay:          currentConfig.PostfixForPickDay,
		UnselectableDaysBeforeTime: timestamppb.New(currentConfig.UnselectableDaysBeforeTime),
		UnselectableDaysAfterTime:  timestamppb.New(currentConfig.UnselectableDaysAfterTime),
		UnselectableDays:           convertTimeMapToProtoMap(currentConfig.UnselectableDays),
	}

	return result
}

func convertTimeMapToProtoMap(tm map[time.Time]struct{}) map[string]*emptypb.Empty {
	protoMap := make(map[string]*emptypb.Empty, len(tm))
	for key := range tm {
		protoMap[key.Format(time.RFC3339)] = &emptypb.Empty{}
	}
	return protoMap
}

func mapToConfigCallbacks(
	req *telegramCalendarPb.NewSettingsRequest,
) (resultCallbacks []func(calendarGenerator.KeyboardGenerator) calendarGenerator.KeyboardGenerator, err error) {
	if req == nil {
		return resultCallbacks, nil
	}

	if req.GetYearsBackForChoose().GetForceChoice() || req.GetYearsBackForChoose().GetYearsBackForChoose() != 0 {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ChangeYearsBackForChoose(int(req.GetYearsBackForChoose().GetYearsBackForChoose())),
		)
	}

	if req.GetYearsForwardForChoose().GetForceChoice() || req.GetYearsForwardForChoose().GetYearsForwardForChoose() != 0 {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ChangeYearsForwardForChoose(int(req.GetYearsForwardForChoose().GetYearsForwardForChoose())),
		)
	}

	days := req.GetDayNames().GetDayNames()

	if len(days) > 0 {
		if len(days) == 7 {
			if errValidate := validateDaysOrMonthsOption(
				days,
				func(d *telegramCalendarPb.NewSettingsRequest_DayName) string { return d.GetDayName() },
				req.GetDayNames().GetForceChoice(),
			); errValidate != nil {
				return resultCallbacks, fmt.Errorf("validateDays error: %w", err)
			}
			var dayNames [7]string
			for i := range dayNames {
				dayNames[i] = days[i].GetDayName()
			}
			resultCallbacks = append(resultCallbacks,
				calendarGenerator.ChangeDaysNames(dayNames),
			)
		} else {
			return resultCallbacks, fmt.Errorf("unexpected number of day names: got %d, want 7", len(days))
		}
	}

	months := req.GetMonthNames().GetMonthNames()
	if len(months) > 0 {
		if len(months) == 12 {
			if errValidate := validateDaysOrMonthsOption(
				months,
				func(m *telegramCalendarPb.NewSettingsRequest_MonthName) string { return m.GetMonthName() },
				req.GetMonthNames().GetForceChoice(),
			); errValidate != nil {
				return resultCallbacks, fmt.Errorf("validateMonths error: %w", err)
			}
			var monthNames [12]string
			for i := range monthNames {
				monthNames[i] = months[i].GetMonthName()
			}
			resultCallbacks = append(resultCallbacks,
				calendarGenerator.ChangeMonthNames(monthNames),
			)
		} else {
			return resultCallbacks, fmt.Errorf("unexpected number of month names: got %d, want 12", len(months))
		}
	}

	if req.GetHomeButtonForBeauty().GetForceChoice() || req.GetHomeButtonForBeauty().GetHomeButtonForBeauty() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ChangeHomeButtonForBeauty(req.GetHomeButtonForBeauty().GetHomeButtonForBeauty()),
		)
	}

	if req.GetPrefixForCurrentDay().GetForceChoice() || req.GetPrefixForCurrentDay().GetPrefixForCurrentDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePrefixForCurrentDay(req.GetPrefixForCurrentDay().GetPrefixForCurrentDay()),
			),
		)
	}

	if req.GetPostfixForCurrentDay().GetForceChoice() || req.GetPostfixForCurrentDay().GetPostfixForCurrentDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePostfixForCurrentDay(req.GetPostfixForCurrentDay().GetPostfixForCurrentDay()),
			),
		)
	}

	if req.GetPrefixForNonSelectedDay().GetForceChoice() || req.GetPrefixForNonSelectedDay().GetPrefixForNonSelectedDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePrefixForNonSelectedDay(req.GetPrefixForNonSelectedDay().GetPrefixForNonSelectedDay()),
			),
		)
	}

	if req.GetPostfixForNonSelectedDay().GetForceChoice() || req.GetPostfixForNonSelectedDay().GetPostfixForNonSelectedDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePostfixForNonSelectedDay(req.GetPostfixForNonSelectedDay().GetPostfixForNonSelectedDay()),
			),
		)
	}

	if req.GetPrefixForPickDay().GetForceChoice() || req.GetPrefixForPickDay().GetPrefixForPickDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePrefixForPickDay(req.GetPrefixForPickDay().GetPrefixForPickDay()),
			),
		)
	}

	if req.GetPostfixForPickDay().GetForceChoice() || req.GetPostfixForPickDay().GetPostfixForPickDay() != "" {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangePostfixForPickDay(req.GetPostfixForPickDay().GetPostfixForPickDay()),
			),
		)
	}

	unselectableDaysBeforeTime := req.GetUnselectableDaysBeforeTime().GetUnselectableDaysBeforeTime().AsTime()
	if req.GetUnselectableDaysBeforeTime().GetForceChoice() ||
		(unselectableDaysBeforeTime != defaultProtoTime && unselectableDaysBeforeTime != unixEpochTime && !unselectableDaysBeforeTime.IsZero()) {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangeUnselectableDaysBeforeDate(unselectableDaysBeforeTime),
			),
		)
	}

	unselectableDaysAfterTime := req.GetUnselectableDaysAfterTime().GetUnselectableDaysAfterTime().AsTime()
	if req.GetUnselectableDaysAfterTime().GetForceChoice() ||
		(unselectableDaysAfterTime != defaultProtoTime && unselectableDaysAfterTime != unixEpochTime && !unselectableDaysAfterTime.IsZero()) {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangeUnselectableDaysAfterDate(unselectableDaysAfterTime),
			),
		)
	}

	if req.GetUnselectableDays().GetForceChoice() || len(req.GetUnselectableDays().GetUnselectableDays()) > 0 {
		unselectableDays := make(map[time.Time]struct{}, len(req.GetUnselectableDays().GetUnselectableDays()))
		for d := range req.GetUnselectableDays().GetUnselectableDays() {
			parsedTime, errParse := time.Parse(time.RFC3339, d)
			if errParse != nil {
				return resultCallbacks, fmt.Errorf("can't parce unselectable days, expect RFC3339 format. Error: %w", errParse)
			}
			unselectableDays[parsedTime] = struct{}{}
		}
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangeUnselectableDays(unselectableDays),
			),
		)
	}

	return resultCallbacks, nil
}

func validateDaysOrMonthsOption[T any](sl []T, getName func(T) string, forceChoice bool) error {
	for i, s := range sl {
		if !forceChoice && getName(s) == "" {
			return fmt.Errorf("unexpected zero value set for item number %v with an active force flag in %v", i, sl)
		}
	}
	return nil
}
