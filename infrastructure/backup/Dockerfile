FROM postgres:17-alpine

ARG TARGETARCH=amd64

RUN apk add --no-cache curl bash \
    && curl -fsSL "https://dl.min.io/client/mc/release/linux-${TARGETARCH}/mc" \
        -o /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc

COPY backup.sh    /usr/local/bin/backup.sh
COPY restore.sh   /usr/local/bin/restore.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /usr/local/bin/backup.sh /usr/local/bin/restore.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
