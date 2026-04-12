FROM golang:1.26-alpine AS builder

COPY services/lib/ /services/lib/

WORKDIR /services/auth-service

COPY services/auth-service/go.mod services/auth-service/go.sum* ./
RUN go mod download

COPY services/auth-service/ .

RUN go build -ldflags="-s -w" -o /out/server  ./cmd/server
RUN go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /out/server  ./server
COPY --from=builder /out/migrate ./migrate
COPY services/auth-service/migrations/ ./migrations/

EXPOSE 50051

CMD ["sh", "-c", "./migrate up && ./server"]
