FROM golang:1.24-alpine

WORKDIR /app
RUN apk add --no-cache git

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o iot-server ./cmd/web

EXPOSE 8080
CMD ["./iot-server"]
