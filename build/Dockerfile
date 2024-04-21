FROM golang:alpine AS builder

WORKDIR /app

COPY . /app

RUN go build -o app main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/app .
CMD ["./app"]