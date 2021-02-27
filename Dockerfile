# Dockerfile for creating a statically-linked Rust application using docker's
# multi-stage build feature. This also leverages the docker build cache to avoid
# re-downloading dependencies if they have not changed.
FROM rust:1.50.0-slim AS build

RUN apt-get update && \
 apt-get install -y \
  zip \
  build-essential \
  musl-tools \
  pkg-config \
  libssl-dev

ENV BUILD_DIR=/build \
    OUTPUT_DIR=/output \
    RUST_BACKTRACE=1 \
    RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    PREFIX=/toolchain \
    MUSL_VERSION=1.1.22 \
    BUILD_TARGET=x86_64-unknown-linux-musl

RUN mkdir -p /usr/local/cargo/bin \
    && mkdir -p $BUILD_DIR \
    && mkdir -p $OUTPUT_DIR \
    && mkdir -p $PREFIX

WORKDIR $PREFIX

ADD http://www.musl-libc.org/releases/musl-$MUSL_VERSION.tar.gz .
RUN tar -xvzf musl-$MUSL_VERSION.tar.gz \
    && cd musl-$MUSL_VERSION \
    && ./configure --prefix=$PREFIX \
    && make install \
    && cd ..

ENV CC=$PREFIX/bin/musl-gcc \
    C_INCLUDE_PATH=$PREFIX/include/ \
    CPPFLAGS=-I$PREFIX/include \
    LDFLAGS=-L$PREFIX/lib

WORKDIR $BUILD_DIR

RUN rustup self update && rustup update
RUN rustup target add $BUILD_TARGET

WORKDIR /usr/src

# Create a dummy project and build the app's dependencies.
# If the Cargo.toml or Cargo.lock files have not changed,
# we can use the docker build cache and skip these (typically slow) steps.
RUN USER=root cargo new smyklot
WORKDIR /usr/src/smyklot
COPY Cargo.toml Cargo.lock ./
RUN cargo build --release

# Copy the source and build the application.
COPY src ./src
RUN cargo install --target x86_64-unknown-linux-musl --path .

# Copy the statically-linked binary into a scratch container.
FROM scratch
COPY --from=build /usr/local/cargo/bin/smyklot .
USER 1000
CMD ["./smyklot"]
