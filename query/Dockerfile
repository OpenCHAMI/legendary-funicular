# SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

ARG TARGETPLATFORM

# With GoReleaser dockers_v2, binaries are available under $TARGETPLATFORM/.
COPY $TARGETPLATFORM/openchami-logq-query /app/openchami-logq-query

# Create a non-root user
RUN addgroup -g 1000 query && \
    adduser -D -u 1000 -G query query && \
    chown -R query:query /app

USER query

ENTRYPOINT ["/app/openchami-logq-query"]
