FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o fyne-client ./cmd/fyne-client/main.go
CMD ["./fyne-client"]