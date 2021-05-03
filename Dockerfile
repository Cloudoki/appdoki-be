# BUILD STAGE
FROM golang:1.16-alpine as builder

ENV GO111MODULE=on
RUN apk update && apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o appdokibin .

# FINAL STAGE
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/appdokibin .
COPY --from=builder /app/swaggerui ./swaggerui
COPY --from=builder /app/migrations ./migrations

EXPOSE 4000
CMD ./appdokibin