FROM golang:1.20-alpine AS builder
WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
