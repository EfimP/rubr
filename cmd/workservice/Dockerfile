FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o workservice ./cmd/workservice/main.go
EXPOSE 50054
CMD ["./workservice"]