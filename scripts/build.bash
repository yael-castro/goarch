#!/bin/bash

# Variables for only read
runtime="github.com/yael-castro/goarch/internal/runtime"
commit=$(git log --pretty=format:'%h' -n 1) || echo 'unknown'

# Command arguments
subcommand="$1"
shift

ldflags=""
options=""

function build() {
    cd "./cmd/$binary" || exit

    if ! go mod tidy
    then
      exit 1
    fi

    if ! go build \
      -o ../../build/ \
      -tags "$tags" \
      -ldflags "$ldflags" \
      "$options"
    then
      exit 1
    fi

    cd ../../

    echo "MD5 checksum: $(md5sum "build/$binary")"
    echo "Success build"
    exit
}


if [ "$subcommand" = "relay" ]; then
  ldflags="-X $runtime.GitCommit=$commit"
  ldflags+=' -extldflags "-static" -linkmode external -w -s '

  binary="users-relay"
  tags="relay,musl"

  printf "\nBuilding CLI in \"build\" directory\n"
  CGO_ENABLED=1 build
fi

if [ "$subcommand" = "http" ]; then
  ldflags="-X $runtime.GitCommit=$commit"
  binary="users-http"
  tags="http"

  printf "\nBuilding API REST in \"build\" directory\n"
  CGO_ENABLED=0 build
fi

echo "Invalid subcommand: $subcommand"
exit 1