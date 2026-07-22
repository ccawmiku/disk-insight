FROM node:24.15.0-alpine3.23 AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts
COPY web/ ./
RUN npm run build

FROM golang:1.26.5-alpine3.23 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd/ ./cmd/
COPY internal/ ./internal/
ARG VERSION=v1.0.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/disk-insight ./cmd/disk-insight

FROM alpine:3.23.5
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 -S disk-insight \
    && adduser -u 10001 -S -D -H -G disk-insight disk-insight \
    && mkdir -p /var/lib/disk-insight /opt/disk-insight/web \
    && chown -R disk-insight:disk-insight /var/lib/disk-insight /opt/disk-insight
COPY --from=go-build /out/disk-insight /usr/local/bin/disk-insight
COPY --from=web-build /src/web/dist /opt/disk-insight/web

ENV DISK_INSIGHT_ADDRESS=:8080 \
    DISK_INSIGHT_DATABASE=/var/lib/disk-insight/disk-insight.db \
    DISK_INSIGHT_WEB=/opt/disk-insight/web \
    DISK_INSIGHT_ROOTS=/data::Data
USER 10001:10001
EXPOSE 8080
VOLUME ["/var/lib/disk-insight"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD wget -qO- http://127.0.0.1:8080/api/v1/health >/dev/null || exit 1
ENTRYPOINT ["/bin/nice", "-n", "10", "/usr/local/bin/disk-insight"]
