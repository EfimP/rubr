FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o superaccservice ./cmd/superaccservice/main.go
EXPOSE 50052
CMD ["./superaccservice"]