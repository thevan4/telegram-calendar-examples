package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mymmrac/telego"
	"github.com/thevan4/telegram-calendar-examples/mymmrac_telego/simplest_example/internal/telegram/telego_wrapper"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	botToken := os.Getenv("BOT_TOKEN")
	botHandler := telego_wrapper.MustNewBotHandler(telego_wrapper.MustNewClient(botToken, telego.WithDefaultDebugLogger()))

	go botHandler.Start()
	<-stop
	botHandler.Stop()
}
