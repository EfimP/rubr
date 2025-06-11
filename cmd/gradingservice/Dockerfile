FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gradingservice ./cmd/gradingservice/main.go
EXPOSE 50057
CMD ["./gradingservice"]