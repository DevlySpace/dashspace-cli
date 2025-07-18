FROM alpine:latest

# Install dependencies
RUN apk add --no-cache ca-certificates git

# Add DashSpace CLI
COPY dashspace-linux-amd64 /usr/local/bin/dashspace
RUN chmod +x /usr/local/bin/dashspace

# Create non-root user
RUN addgroup -g 1000 dashspace && \
    adduser -D -s /bin/sh -u 1000 -G dashspace dashspace

USER dashspace
WORKDIR /workspace

ENTRYPOINT ["dashspace"]
CMD ["--help"]