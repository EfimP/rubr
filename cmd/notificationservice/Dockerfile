FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notificationservice ./cmd/notificationservice/main.go
EXPOSE 50056
CMD ["./notificationservice"]