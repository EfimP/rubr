# Этап 1: Сборка с зависимостями
FROM golang:1.24.3 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Сборка для linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gradingservice ./cmd/gradingservice/main.go
RUN chmod +x ./gradingservice

# Этап 2: Финальный образ
FROM golang:1.24.3
WORKDIR /app
COPY --from=builder /app/gradingservice .
CMD ["./gradingservice"]