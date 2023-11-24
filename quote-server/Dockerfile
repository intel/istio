FROM rust:1.74.0 as quote-server-builder

RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    make \
    libprotobuf-dev \
    protobuf-compiler \
    musl-dev \
    wget \
    libssl-dev \
    pkg-config

WORKDIR /app

COPY . .

COPY quote-server /quote-server

RUN cd /quote-server && make build

FROM rust:1.74.0

WORKDIR /app
COPY --from=quote-server-builder /quote-server/target/release/quote_server /app/quote_server

RUN chmod a+x /app/quote_server

# Run the sleep command for demonstration purposes
CMD ["/app/quote_server"]
