# Working as a separate http/grpc microservice

## What is it and why?

Running a fully autonomous service in a container. You don't need to understand golang for this, you can just grab an http contract with swagger or proto for grpc and start using it!

Usage examples with specific languages and frameworks will be posted separately, which doesn't stop you from figuring out the contracts and starting to use the app.

### Contracts
- [proto](https://github.com/thevan4/telegram-calendar-examples/blob/main/proto/telegram_calendar.proto)
- http examples at swagger, just [run it](https://hub.docker.com/r/thevan/telegram-calendar-standalone-go)! 

#### Codegen
```bash
make gen
```

#### Local build and run
```bash
make build-and-run
```

#### Docker run
```bash
make docker-run
```
