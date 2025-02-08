# Initial version copied from this example:
# https://github.com/astral-sh/uv-docker-example/blob/1c593aa32621eacd0125b55bd8d2796b86e8ea73/Dockerfile

# Use a Python image with uv pre-installed
FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim

RUN apt-get update -qq && \
    apt-get install --no-install-recommends -y sqlite3 curl && \
    rm -rf /var/lib/apt/lists /var/cache/apt/archives

ARG LITESTREAM_VERSION=0.3.13
RUN curl https://github.com/benbjohnson/litestream/releases/download/v0.3.13/litestream-v${LITESTREAM_VERSION}-linux-amd64.deb -O -L
RUN dpkg -i litestream-v${LITESTREAM_VERSION}-linux-amd64.deb

# Install the project into `/app`
WORKDIR /app

# Enable bytecode compilation
ENV UV_COMPILE_BYTECODE=1

# Copy from the cache instead of linking since it's a mounted volume
ENV UV_LINK_MODE=copy

# Install the project's dependencies using the lockfile and settings
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --frozen --no-install-project --no-dev

# Then, add the rest of the project source code and install it
# Installing separately from its dependencies allows optimal layer caching
ADD . /app
COPY litestream.yml /etc/litestream.yml
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

RUN export SENTRY_RELEASE=$(git rev-parse HEAD) && \
    echo "SENTRY_RELEASE=$SENTRY_RELEASE" >> /etc/environment


# Place executables in the environment at the front of the path
ENV PATH="/app/.venv/bin:$PATH"

ENTRYPOINT ["bin/entrypoint"]
