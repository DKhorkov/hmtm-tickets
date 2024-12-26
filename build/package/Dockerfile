# https://habr.com/ru/companies/otus/articles/660301/
FROM golang:alpine AS builder
LABEL authors='dkhorkov'

WORKDIR /build

COPY . .

RUN go build -o hmtmtickets ./cmd/hmtmtickets/main.go

FROM alpine AS runner

WORKDIR /app

COPY .env .
COPY --from=builder /build/migrations/ /app/migrations/
COPY --from=builder /build/hmtmtickets /app/hmtmtickets

CMD ["./hmtmtickets"]
