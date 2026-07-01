FROM golang:1.25.7-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o blog .

FROM alpine:3.21.3

WORKDIR /app

RUN apk add --no-cache tzdata wget

COPY --from=builder /app/blog .
COPY --from=builder /app/templates ./templates

EXPOSE 8083

CMD ["./blog", "server"]
