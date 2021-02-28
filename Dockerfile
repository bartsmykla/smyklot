FROM ghcr.io/bartsmykla/rust-musl:0.1.0 AS build

WORKDIR /usr/src

# Create a dummy project and build the app's dependencies.
# If the Cargo.toml or Cargo.lock files have not changed,
# we can use the docker build cache and skip these (typically slow) steps.
RUN USER=root cargo new smyklot
WORKDIR /usr/src/smyklot
COPY Cargo.toml Cargo.lock src target ./
RUN cargo install --target x86_64-unknown-linux-musl --path .

# Copy the statically-linked binary into a scratch container.
FROM scratch
COPY --from=build /usr/local/cargo/bin/smyklot .
USER 1000
CMD ["./smyklot"]
