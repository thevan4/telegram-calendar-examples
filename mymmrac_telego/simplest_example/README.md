# Works as a full service

The easiest way to use [telegram-calendar](https://github.com/thevan4/telegram-calendar)  is with [telego](https://github.com/mymmrac/telego).

## Run

1. Set the bot token in the environment variables - `BOT_TOKEN`. If the token is not set or is not valid, an error will occur: `panic: at MustNewClient: telego: invalid token` 
2. Run `go run cmd/main.go` (it is assumed that you are in the `telegram-calendar-examples/mymmrac_telego/simplest_example` folder).