FROM golang:1.22.2 as builder
WORKDIR /app
COPY ../../.. .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd
FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /app/main /main
EXPOSE 50051 8080
ENTRYPOINT ["/main"]
