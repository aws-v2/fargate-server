# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mini-fargate cmd/api/main.go

# Final stage
FROM alpine:3.19

RUN adduser -D -u 1000 appuser

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/mini-fargate .
COPY --from=builder /app/docs ./docs

USER appuser

EXPOSE 8086

CMD ["./mini-fargate"]