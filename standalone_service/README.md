# Working as a separate http/grpc microservice

## What is it and why?

Running a fully autonomous service in a container. You don't need to understand golang for this, you can just grab an http contract with swagger or proto for grpc and start using it!

Usage examples with specific languages and frameworks will be posted separately, which doesn't stop you from figuring out the contracts and starting to use the app.

## Easy start

### Docker compose run

Configuration example.

The section with swagger-ui is optional if you don't need swagger. At the moment you can't send requests directly through swagger, it's only for the example of working with handlers (curl request is generated valid, but to wrong address).

Default swagger url: http://127.0.0.1:8081/swagger (but the requests have to go to an address 127.0.0.1:8080).

```yml
services:
  telegram-calendar-service:
    image: thevan/telegram-calendar-standalone-go
    ports:
      # "destination_out_container:source"
      - "50051:50051" # Port for gRPC
      - "8080:8080" # Port for HTTP
    environment:
      GRPC_PORT: "50051"
      HTTP_PORT: "8080"
      GRPC_DIAL_TIMEOUT: "1s"
      HTTP_DIAL_TIMEOUT: "1s"
      BEARER_TOKEN: "" # insecure, needs to be changed. For grpc header "authorization", for http "Authorization" (it's just the way it is)

  swagger-ui:
    image: swaggerapi/swagger-ui
    restart: unless-stopped
    ports:
      - "8081:8080" # swagger UI will be available at http://localhost:8081/swagger
    volumes:
      - ../../api/swagger:/usr/share/nginx/html/swagger
    environment:
      SWAGGER_JSON: /usr/share/nginx/html/swagger/telegram.calendar.swagger.json
      BASE_URL: /swagger
```

### Example request

```bash
curl -X 'POST' \
  'http://127.0.0.1:8080/generate-calendar' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "callbackPayload": "",
  "currentTime": "2024-01-01T00:00:00.000Z"
}'
```

### Contracts
- [proto](https://github.com/thevan4/telegram-calendar-examples/blob/main/standalone_service/proto/telegram_calendar.proto)
- http examples at swagger, just [run it](https://hub.docker.com/r/thevan/telegram-calendar-standalone-go)! 

# Codegen (only for project work, not for depoy)
```bash
make gen
```

## Local build and run
```bash
make build-and-run
```

## Docker run
```bash
make docker-run
```
