# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

# install dependencies ก่อน (cache layer)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o arise-bank ./main.go

# Stage 2: Run (image เล็กสุด)
FROM alpine:3.21

WORKDIR /app

# สำหรับ health check และ TLS
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/arise-bank .

EXPOSE 8080

CMD ["./arise-bank"]
