FROM golang:1.24.3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o workassignmentservice ./cmd/workassignmentservice/main.go
EXPOSE 50054
CMD ["./workassignmentservice"]