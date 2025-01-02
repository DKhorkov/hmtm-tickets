# https://habr.com/ru/companies/otus/articles/660301/
FROM golang:alpine AS builder
LABEL authors='dkhorkov'

WORKDIR /build

COPY . .

RUN go build -o server ./cmd/server/server.go

FROM alpine AS runner

WORKDIR /app

COPY .env .
COPY --from=builder /build/migrations/ /app/migrations/
COPY --from=builder /build/server /app/server

CMD ["./server"]
