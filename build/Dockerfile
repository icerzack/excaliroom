FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o app main.go

FROM alpine
COPY --from=builder /app/app .
CMD ["./app"]