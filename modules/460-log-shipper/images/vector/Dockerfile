ARG BASE_DEBIAN_BULLSEYE
ARG BASE_RUST

FROM $BASE_RUST as build

RUN apt-get update \
    && apt-get install -yq \
      ca-certificates make bash cmake libclang1-9 llvm-9 libsasl2-dev librdkafka-dev

WORKDIR /vector
RUN git clone --depth 1 --branch v0.24.2 https://github.com/vectordotdev/vector.git \
    && cd vector

# Download and cache dependencies
WORKDIR /vector/vector
RUN cargo fetch

RUN cargo build \
    --release \
    -j $(($(nproc) /2)) \
    --offline \
    --no-default-features \
    --features "api,api-client,enrichment-tables,sources-host_metrics,sources-internal_metrics,sources-file,sources-kubernetes_logs,transforms,sinks-prometheus,sinks-blackhole,sinks-elasticsearch,sinks-file,sinks-loki,sinks-socket,sinks-console,sinks-vector,sinks-kafka,unix,rdkafka?/gssapi-vendored,vrl-cli" \
    && strip target/release/vector

FROM $BASE_DEBIAN_BULLSEYE
RUN mkdir -p /etc/vector \
    && apt-get update \
    && apt-get install -yq ca-certificates tzdata inotify-tools gettext procps \
    && rm -rf /var/cache/apt/archives/*
COPY --from=build /vector/vector/target/release/vector /usr/bin/vector
COPY reloader /usr/bin/reloader
ENTRYPOINT ["/usr/bin/vector"]
