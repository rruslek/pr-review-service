FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY static/ ./static/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pr-review-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/pr-review-service .
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./pr-review-service"]

