# Works as a full service

The easiest way to use [telegram-calendar](https://github.com/thevan4/telegram-calendar)  is with [telego](https://github.com/mymmrac/telego).

## Run

1. Start docker container for [standalone generator](https://github.com/thevan4/telegram-calendar-examples/tree/main/standalone_service).
2. Set the bot token in the environment variables: `BOT_TOKEN`. If the token is not set or is not valid, an error will occur: `panic: at MustNewClient: telego: invalid token`
3. Set grpc address in the environment variables for generator service: `CALENDAR_GENERATOR_GRPC_ADDR`). If the token is not set or is not valid, an error will occur: `at MustNewClientForStandaloneService: addr not set`
4. Run `go run cmd/main.go` (it is assumed that you are in the `telegram-calendar-examples/mymmrac_telego/with_standalone_service_example` folder).
