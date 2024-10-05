# Build stage
FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/simplebank/main.go

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY internal/db/migration ./internal/db/migration

EXPOSE 8080
CMD ["./main"]
ENTRYPOINT [ "/app/start.sh" ]