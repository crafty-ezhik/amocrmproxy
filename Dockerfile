# Stage 1 — Build
FROM golang:1.24.2-alpine AS builder

RUN apk --no-cache add bash gcc gettext musl-dev

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY ./ ./


RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/app ./cmd/main.go

# Stage 2 — Final Image
FROM alpine:latest AS runner

COPY --from=builder /app/bin/app /

COPY ./config.json ./

EXPOSE 8089

CMD ["/app"]