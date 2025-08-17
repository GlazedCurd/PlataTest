FROM golang:1.25 AS builder

WORKDIR /app/worker

COPY go.mod go.sum ./
RUN go mod download

COPY ../../ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/worker/main.go

FROM alpine:latest

COPY --from=builder /app/worker/main /main

CMD ["/main"]