# Этап 1: Сборка с зависимостями
FROM golang:1.24.3 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Установка системных зависимостей для Fyne (OpenGL и X11) для сборки
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    && rm -rf /var/lib/apt/lists/*
COPY . .
# Явно указываем архитектуру linux/amd64
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o fyne-client ./cmd/fyne-client/main.go
RUN chmod +x ./fyne-client

# Этап 2: Финальный образ с графическими библиотеками
FROM golang:1.24.3
WORKDIR /app
# Установка runtime-зависимостей для Fyne
RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    libx11-dev \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/fyne-client .
CMD ["./fyne-client"]