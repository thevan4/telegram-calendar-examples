package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mymmrac/telego"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/telego_wrapper"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// You can set your own context, for more flexible control of graceful shutdown.
	ctx := context.Background()
	botToken := os.Getenv("BOT_TOKEN")
	calendarStandaloneServiceAddr := os.Getenv("CALENDAR_GENERATOR_GRPC_ADDR")
	botHandler := telego_wrapper.MustNewBotHandler(ctx, telego_wrapper.MustNewClient(botToken, telego.WithDefaultDebugLogger()),
		calendarStandaloneServiceAddr)

	go botHandler.Start()
	<-stop
	botHandler.Stop()
}
