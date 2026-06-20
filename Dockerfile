FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ticketer ./cmd/ticketer

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /data
COPY --from=builder /build/ticketer /usr/local/bin/ticketer
EXPOSE 8080
CMD ["ticketer"]
