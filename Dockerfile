# syntax=docker/dockerfile:1

FROM scratch

ARG TARGETPLATFORM

LABEL org.opencontainers.image.source="https://github.com/bartsmykla/smyklot"
LABEL org.opencontainers.image.description="Automated PR approvals and merges based on CODEOWNERS"
LABEL org.opencontainers.image.licenses="MIT"

# Copy pre-built binary from GoReleaser
COPY ${TARGETPLATFORM}/smyklot /smyklot

ENTRYPOINT ["/smyklot"]
