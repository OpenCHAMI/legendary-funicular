# SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY openchami-logq-compactor /app/openchami-logq-compactor

# Create a non-root user
RUN addgroup -g 1000 compactor && \
    adduser -D -u 1000 -G compactor compactor && \
    chown -R compactor:compactor /app

USER compactor

ENTRYPOINT ["/app/openchami-logq-compactor"]
