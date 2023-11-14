FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-parser .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

ENV TELEGRAM_BOT_TOKEN=
ENV TELEGRAM_BOT_ADMIN_ID=
ENV ENVIRONMENT=production

COPY config/config.yml /app/config/config.yml
COPY --from=builder /app/telegram-parser .

CMD [ "./telegram-parser" ]