# ===== Build stage =====
FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binary statis (mssql driver murni Go, tidak butuh CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server .

# ===== Runtime stage =====
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta

COPY --from=build /app/server /app/server

ENV APP_HOST=0.0.0.0 \
    APP_PORT=5500

EXPOSE 5500

ENTRYPOINT ["/app/server"]
