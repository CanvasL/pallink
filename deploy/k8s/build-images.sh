#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE_TAG="${IMAGE_TAG:-dev}"

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

build_image() {
  local image_name="$1"
  local dockerfile_path="$2"

  printf '[k8s-build] %s\n' "pallink/${image_name}:${IMAGE_TAG}"
  docker build \
    -f "${dockerfile_path}" \
    -t "pallink/${image_name}:${IMAGE_TAG}" \
    "${ROOT_DIR}"
}

require docker

build_image gateway-api "${ROOT_DIR}/deploy/docker/gateway-api.Dockerfile"
build_image user-rpc "${ROOT_DIR}/deploy/docker/user-rpc.Dockerfile"
build_image activity-rpc "${ROOT_DIR}/deploy/docker/activity-rpc.Dockerfile"
build_image notification-rpc "${ROOT_DIR}/deploy/docker/notification-rpc.Dockerfile"
build_image im-rpc "${ROOT_DIR}/deploy/docker/im-rpc.Dockerfile"
build_image audit-worker "${ROOT_DIR}/deploy/docker/audit-worker.Dockerfile"
build_image postgres "${ROOT_DIR}/deploy/k8s/docker/postgres.Dockerfile"
build_image swagger-ui "${ROOT_DIR}/deploy/k8s/docker/swagger-ui.Dockerfile"
build_image prometheus "${ROOT_DIR}/deploy/k8s/docker/prometheus.Dockerfile"
build_image grafana "${ROOT_DIR}/deploy/k8s/docker/grafana.Dockerfile"
