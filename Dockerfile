FROM busybox:1.25-glibc

WORKDIR /app
ENTRYPOINT ["/app/tigerblood"]

RUN addgroup -g 10001 app && \
    adduser -G app -u 10001 -D -h /app -s /sbin/nologin app

COPY version.json /app/version.json
COPY tigerblood /app/tigerblood

USER app
