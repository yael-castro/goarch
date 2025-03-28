FROM golang:tip-alpine3.21 AS builder

WORKDIR /app

# Install "make" command
RUN apk add --no-cache make

# Install GCC
RUN apk add --no-cache build-base

# Git installation.
#  It is required to extract the hash of the last commit to use it as the version of the compiled binaries.
RUN apk add --no-cache git

COPY . .

# Compiling only whats is required
ARG TAG
RUN make "$TAG"

FROM alpine:3.21.3

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/build .

# Executing compiled binary
CMD "./$EXECUTABLE"