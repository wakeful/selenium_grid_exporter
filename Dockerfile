FROM golang:1.14.2-alpine3.11 AS builder
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

ENV USER=appuser
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/github.com/wakeful/selenium_grid_exporter
COPY . .

RUN go mod download && \
  go mod verify && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags="-w -s" -o /go/bin/selenium_grid_exporter

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/selenium_grid_exporter /go/bin/selenium_grid_exporter
USER appuser:appuser
EXPOSE 8080
ENTRYPOINT ["/go/bin/selenium_grid_exporter"]

