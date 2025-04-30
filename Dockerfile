FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/web

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/server .

COPY templates ./templates
COPY static    ./static

COPY .env .env

EXPOSE 3000

RUN addgroup -S app && adduser -S -G app app
USER app

# Command
CMD ["./server"]
