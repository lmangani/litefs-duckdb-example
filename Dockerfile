# Build our application using a Go builder.
FROM golang:1.23 AS builder

WORKDIR /src/litefs-example
COPY . .
RUN CGO_ENABLED=1 go build -buildvcs=false -tags osusergo,netgo -o /usr/local/bin/litefs-example ./cmd/litefs-example


# Our final Docker image stage starts here.
FROM debian:bookworm
ARG LITEFS_CONFIG=litefs.yml

# Copy binaries from the previous build stages.
COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs
COPY --from=builder /usr/local/bin/litefs-example /usr/local/bin/litefs-example

# Copy the possible LiteFS configurations.
ADD fly-io-config/etc/litefs.yml /tmp/litefs.yml
ADD docker-config/etc/litefs.static-lease.yml /tmp/litefs.static-lease.yml

# Move the appropriate LiteFS config file to /etc/ (this one will be
# used by LiteFS). By default this is the config file used on Fly.io,
# but it's set appropriately to other files for the docker setup in
# docker-compose.yml
RUN cp /tmp/$LITEFS_CONFIG /etc/litefs.yml

# Setup our environment to include FUSE & SQLite. We install ca-certificates
# so we can communicate with the Consul server over HTTPS. cURL is added so
# we can call our HTTP endpoints for debugging.
RUN apt-get update -y && apt-get install -y ca-certificates fuse3

# Run LiteFS as the entrypoint. After it has connected and sync'd with the
# cluster, it will run the commands listed in the "exec" field of the config.
ENTRYPOINT litefs mount
