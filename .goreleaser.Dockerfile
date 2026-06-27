FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /data
COPY goard /usr/local/bin/goard
EXPOSE 8300
CMD ["goard"]
