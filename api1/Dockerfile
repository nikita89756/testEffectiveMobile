FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build  -o /app/main ./cmd/mainapi/main.go

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/main .
COPY --from=builder /app/internal/storage/postgres/migrations ./internal/storage/postgres/migrations
EXPOSE 8080
CMD ["./main"]