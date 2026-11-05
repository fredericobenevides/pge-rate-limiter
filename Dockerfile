FROM golang:1.27 as build

ARG APP_PATH

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags=viper_bind_struct -o main ./cmd/rate-limiter/main.go

FROM scratch

WORKDIR /app

COPY --from=build /app/main .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["./main"]
