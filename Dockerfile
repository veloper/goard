FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o goard ./cmd/goard && \
    CGO_ENABLED=0 go build -o goardctl ./cmd/goardctl

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /data
COPY --from=builder /build/goard /usr/local/bin/goard
COPY --from=builder /build/goardctl /usr/local/bin/goardctl
ENV GOARD_DB_PATH=/data/goard.db
EXPOSE 8300
CMD ["goard"]
