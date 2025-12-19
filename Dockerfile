FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY .env /app/.env
EXPOSE 8080
CMD ["/app/server"]
