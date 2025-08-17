FROM golang:1.25 AS builder

WORKDIR /app/server

COPY go.mod go.sum ./
RUN go mod download

COPY ../../ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

FROM alpine:latest

COPY --from=builder /app/server/main /main

CMD ["/main"]