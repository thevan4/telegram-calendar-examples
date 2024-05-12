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

func mapGenerateCalendarKeyboardToResponse(
	generatedCalendarKeyboard models.GenerateCalendarKeyboardResponse,
) *telegramCalendarPb.GenerateCalendarResponse {
	// mapping InlineKeyboardMarkup
	inlineKeyboardMarkupResponse := &telegramCalendarPb.InlineKeyboardMarkup{
		InlineKeyboard: make([]*telegramCalendarPb.InlineKeyboardMarkup_InlineKeyboardRow, len(generatedCalendarKeyboard.InlineKeyboardMarkup.InlineKeyboard)),
	}
	for i, row := range generatedCalendarKeyboard.InlineKeyboardMarkup.InlineKeyboard {
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

	isUnselectableDay := generatedCalendarKeyboard.IsUnselectableDay
	// mapping result
	result := &telegramCalendarPb.GenerateCalendarResponse{
		InlineKeyboardMarkup: inlineKeyboardMarkupResponse,
		IsUnselectableDay:    &isUnselectableDay,
	}

	if !generatedCalendarKeyboard.SelectedDay.IsZero() {
		result.SelectedDay = timestamppb.New(generatedCalendarKeyboard.SelectedDay)
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
		UnselectableDaysBeforeTime: currentConfig.UnselectableDaysBeforeTime.Format(time.RFC3339),
		UnselectableDaysAfterTime:  currentConfig.UnselectableDaysAfterTime.Format(time.RFC3339),
		UnselectableDays:           convertTimeMapToProtoMap(currentConfig.UnselectableDays),
		Timezone:                   currentConfig.Timezone.String(),
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

// A better approach is to take all conversions to the layer above, and perform the conversion from the transport layer
// to the domain layer together with the request validation.
// But it's an additional cost, that's why it's fine for now.
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

	var unselectableDaysBeforeTime time.Time
	var errParseDBT error
	if req.GetUnselectableDaysBeforeTime().GetUnselectableDaysBeforeTime() != "" {
		unselectableDaysBeforeTime, errParseDBT = time.Parse(time.RFC3339, req.GetUnselectableDaysBeforeTime().GetUnselectableDaysBeforeTime())
		if errParseDBT != nil {
			return resultCallbacks, fmt.Errorf("parse unselectable days before time error: %v", errParseDBT)
		}
	}
	if req.GetUnselectableDaysBeforeTime().GetForceChoice() ||
		!unselectableDaysBeforeTime.IsZero() {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangeUnselectableDaysBeforeDate(unselectableDaysBeforeTime),
			),
		)
	}

	var unselectableDaysAfterTime time.Time
	var errParseDAT error
	if req.GetUnselectableDaysAfterTime().GetUnselectableDaysAfterTime() != "" {
		unselectableDaysAfterTime, errParseDAT = time.Parse(time.RFC3339, req.GetUnselectableDaysAfterTime().GetUnselectableDaysAfterTime())
		if errParseDAT != nil {
			return resultCallbacks, fmt.Errorf("parse unselectable days before time error: %v", errParseDAT)
		}
	}

	if req.GetUnselectableDaysAfterTime().GetForceChoice() ||
		!unselectableDaysAfterTime.IsZero() {
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

	var newTimezone *time.Location
	var errParseTimezone error
	if req.GetTimezone().GetTimezone() != "" {
		newTimezone, errParseTimezone = time.LoadLocation(req.GetTimezone().GetTimezone())
		if errParseTimezone != nil {
			return resultCallbacks, fmt.Errorf("parse new timezone error: %v", errParseTimezone)
		}
	}

	// new timezone may be nil, this is normal, it will be set to UTC
	if req.GetTimezone().GetForceChoice() || newTimezone != nil {
		resultCallbacks = append(resultCallbacks,
			calendarGenerator.ApplyNewOptionsForButtonsTextWrapper(
				telegramDayButtonFormer.ChangeTimezone(newTimezone),
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
