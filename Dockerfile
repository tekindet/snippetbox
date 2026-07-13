FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o sniper cmd/web/*

FROM alpine:3.19 AS production

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/sniper .
COPY --from=builder /app/ui ./ui
COPY --from=builder /app/migrations ./migrations

EXPOSE 5000

CMD ["./sniper"]
