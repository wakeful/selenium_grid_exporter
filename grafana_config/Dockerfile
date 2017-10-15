FROM alpine:3.6

WORKDIR /grafana_config
COPY . .

RUN apk add --no-cache curl
RUN chmod +x /grafana_config/wait.sh

ENTRYPOINT ["/grafana_config/wait.sh"]
