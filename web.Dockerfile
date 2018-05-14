FROM golang:1.10

RUN addgroup --gid 10001 app && \
    \
    adduser --gid 10001 \
            --uid 10001 \
            --home /app \
            --gecos '' \
            --shell /sbin/nologin \
            --disabled-password app && \
    \
    apt update && \
    apt -y upgrade && \
    apt-get clean

ADD . /go/src/go.mozilla.org/tigerblood
ADD version.json /app
RUN mkdir -p /app/bin/
COPY bin/run.sh /app/bin/run.sh

RUN go install go.mozilla.org/tigerblood/cmd/tigerblood

USER app
WORKDIR /app
ENTRYPOINT ["/bin/bash", "/app/bin/run.sh"]
EXPOSE 8080
