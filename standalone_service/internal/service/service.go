package service

import (
	"context"
	"fmt"
	"time"

	telegramCalendarPb "github.com/thevan4/telegram-calendar-examples/standalone_service/pkg/telegram-calendar"
	calendarManager "github.com/thevan4/telegram-calendar/manager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TelegramCalendarGRPCService ...
type TelegramCalendarGRPCService struct {
	telegramCalendarPb.UnimplementedCalendarServiceServer
	calendarManager calendarManager.KeyboardManager
}

// NewTelegramCalendarGRPCService ...
func NewTelegramCalendarGRPCService() *TelegramCalendarGRPCService {
	return &TelegramCalendarGRPCService{calendarManager: calendarManager.NewManager()}
}

// GenerateCalendar ...
func (g *TelegramCalendarGRPCService) GenerateCalendar(
	_ context.Context,
	req *telegramCalendarPb.GenerateCalendarRequest,
) (*telegramCalendarPb.GenerateCalendarResponse, error) {
	currentTime, err := time.Parse(time.RFC3339, req.GetCurrentTime())
	if err != nil {
		return new(telegramCalendarPb.GenerateCalendarResponse), fmt.Errorf("parse current time error: %v", err)
	}
	return mapGenerateCalendarKeyboardToResponse(
		g.calendarManager.GenerateCalendarKeyboard(req.GetCallbackPayload(), currentTime),
	), nil
}

// GetSettings ...
func (g *TelegramCalendarGRPCService) GetSettings(
	_ context.Context,
	_ *emptypb.Empty,
) (*telegramCalendarPb.GetSettingsResponse, error) {
	return mapCurrentConfigToResponse(g.calendarManager.GetCurrentConfig()), nil
}

// ApplyNewSettings ...
func (g *TelegramCalendarGRPCService) ApplyNewSettings(
	_ context.Context,
	req *telegramCalendarPb.NewSettingsRequest,
) (*emptypb.Empty, error) {
	newOptions, err := mapToConfigCallbacks(req)
	if err != nil {
		return new(emptypb.Empty), status.Errorf(codes.InvalidArgument, "invalid data in NewSettingsRequest: %v", err)
	} else if len(newOptions) == 0 {
		return new(emptypb.Empty), nil
	}

	g.calendarManager.ApplyNewOptions(newOptions...)
	return new(emptypb.Empty), nil
}
