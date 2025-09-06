FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o questions-vote ./cmd/bot

FROM debian:bookworm-slim
RUN apt-get update -qq && \
    apt-get install --no-install-recommends -y sqlite3 curl ca-certificates && \
    rm -rf /var/lib/apt/lists /var/cache/apt/archives

ARG LITESTREAM_VERSION=0.3.13
RUN curl https://github.com/benbjohnson/litestream/releases/download/v0.3.13/litestream-v${LITESTREAM_VERSION}-linux-amd64.deb -O -L
RUN dpkg -i litestream-v${LITESTREAM_VERSION}-linux-amd64.deb

WORKDIR /app

COPY --from=builder /app/questions-vote /app/questions-vote
COPY litestream.yml /etc/litestream.yml
COPY bin/entrypoint /app/bin/entrypoint
RUN chmod +x /app/bin/entrypoint

RUN export SENTRY_RELEASE=$(git rev-parse HEAD) && \
    echo "SENTRY_RELEASE=$SENTRY_RELEASE" >> /etc/environment

ENTRYPOINT ["bin/entrypoint"]
