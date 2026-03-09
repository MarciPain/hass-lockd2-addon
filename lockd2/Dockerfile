ARG BUILD_FROM=ghcr.io/home-assistant/aarch64-base:latest
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY backend /app
RUN go mod tidy
RUN go mod download
RUN go build -o /lockd2-addon .

FROM $BUILD_FROM

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /lockd2-addon /usr/bin/lockd2-addon
COPY run.sh /run.sh
RUN chmod +x /run.sh

CMD [ "/run.sh" ]
