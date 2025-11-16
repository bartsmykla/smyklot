# syntax=docker/dockerfile:1

FROM scratch

ARG TARGETPLATFORM

# Copy pre-built binary from GoReleaser
COPY ${TARGETPLATFORM}/smyklot /smyklot

ENTRYPOINT ["/smyklot"]
