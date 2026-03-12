#!/bin/bash

set -e

ROOT=$(pwd)

generate_api() {
  service=$1
  api_file="$ROOT/$service/api/$service.api"

  if [ -f "$api_file" ]; then
    echo "Generating API for $service ..."
    goctl api go \
      -api $api_file \
      -dir $ROOT/$service/api
  else
    echo "No api file found for $service"
  fi
}

generate_rpc() {
  service=$1
  proto_file="$ROOT/$service/rpc/$service.proto"

  if [ -f "$proto_file" ]; then
    echo "Generating RPC for $service ..."
    goctl rpc protoc $proto_file \
      --go_out=$ROOT/$service/rpc \
      --go-grpc_out=$ROOT/$service/rpc \
      --zrpc_out=$ROOT/$service/rpc \
      -I=$ROOT \
      -I=$ROOT/include
  else
    echo "No proto file found for $service"
  fi
}

case $1 in
  api)
    generate_api $2
    ;;
  rpc)
    generate_rpc $2
    ;;
  all)
    for dir in */ ; do
      service=${dir%/}
      generate_api $service
      generate_rpc $service
    done
    ;;
  *)
    echo "Usage:"
    echo "./gen.sh api user"
    echo "./gen.sh rpc user"
    echo "./gen.sh all"
    ;;
esac