#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_NAME="${KIND_CLUSTER_NAME:-pallink-local}"
IMAGE_TAG="${IMAGE_TAG:-dev}"
NAMESPACE="${K8S_NAMESPACE:-pallink}"
KUSTOMIZE_DIR="${ROOT_DIR}/deploy/k8s/overlays/local"
KIND_CONFIG="${KUSTOMIZE_DIR}/kind-config.yaml"
KUBE_CONTEXT="kind-${CLUSTER_NAME}"
INGRESS_NGINX_MANIFEST="${INGRESS_NGINX_MANIFEST:-https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.14.1/deploy/static/provider/kind/deploy.yaml}"

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

load_image() {
  local image_name="$1"
  printf '[k8s-kind] load %s\n' "pallink/${image_name}:${IMAGE_TAG}"
  kind load docker-image --name "${CLUSTER_NAME}" "pallink/${image_name}:${IMAGE_TAG}"
}

normalize_proxy_value() {
  local value="${1:-}"

  if [[ -z "${value}" ]]; then
    return 0
  fi

  if [[ "${value}" =~ ^http://(127\.0\.0\.1|localhost)(:([0-9]+))?$ ]]; then
    local port="${BASH_REMATCH[3]}"
    if [[ -n "${port}" ]]; then
      printf 'http://host.docker.internal:%s\n' "${port}"
    else
      printf 'http://host.docker.internal\n'
    fi
    return 0
  fi

  if [[ "${value}" =~ ^https://(127\.0\.0\.1|localhost)(:([0-9]+))?$ ]]; then
    local port="${BASH_REMATCH[3]}"
    if [[ -n "${port}" ]]; then
      printf 'https://host.docker.internal:%s\n' "${port}"
    else
      printf 'https://host.docker.internal\n'
    fi
    return 0
  fi

  printf '%s\n' "${value}"
}

kind_create_cluster() {
  local env_args=()
  local var normalized

  for var in HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NO_PROXY no_proxy; do
    if [[ -n "${!var:-}" ]]; then
      if [[ "${var}" == "NO_PROXY" || "${var}" == "no_proxy" ]]; then
        normalized="${!var}"
      else
        normalized="$(normalize_proxy_value "${!var}")"
      fi
      env_args+=("${var}=${normalized}")
      if [[ "${normalized}" != "${!var}" ]]; then
        printf '[k8s-kind] rewrite %s for kind node env: %s -> %s\n' "${var}" "${!var}" "${normalized}"
      fi
    fi
  done

  env "${env_args[@]}" kind create cluster --name "${CLUSTER_NAME}" --config "${KIND_CONFIG}"
}

rollout() {
  local deployment_name="$1"
  printf '[k8s-kind] wait deployment/%s\n' "${deployment_name}"
  kubectl --context "${KUBE_CONTEXT}" -n "${NAMESPACE}" rollout status "deployment/${deployment_name}" --timeout=300s
}

restart() {
  local deployment_name="$1"
  kubectl --context "${KUBE_CONTEXT}" -n "${NAMESPACE}" rollout restart "deployment/${deployment_name}" >/dev/null
}

dump_ingress_debug() {
  printf '\n[k8s-kind] ingress-nginx diagnostics\n' >&2
  kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx get pods,jobs 2>&1 >&2 || true
  kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx get events --sort-by=.lastTimestamp 2>&1 | tail -n 30 >&2 || true
  kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx describe job ingress-nginx-admission-create >&2 || true
  kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx describe job ingress-nginx-admission-patch >&2 || true
}

wait_for_job_if_exists() {
  local job_name="$1"

  if ! kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx get job "${job_name}" >/dev/null 2>&1; then
    return 0
  fi

  kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx wait \
    --for=condition=complete "job/${job_name}" \
    --timeout=300s
}

wait_for_secret() {
  local secret_name="$1"
  local attempt

  for attempt in $(seq 1 60); do
    if kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx get secret "${secret_name}" >/dev/null 2>&1; then
      return 0
    fi

    sleep 2
  done

  printf 'secret %s did not become ready in time\n' "${secret_name}" >&2
  return 1
}

wait_for_admission_endpoint() {
  local attempt

  for attempt in $(seq 1 60); do
    if kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx get endpoints ingress-nginx-controller-admission \
      -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null | grep -q '[0-9]'; then
      return 0
    fi

    sleep 2
  done

  printf 'ingress-nginx admission endpoint did not become ready in time\n' >&2
  return 1
}

install_ingress_nginx() {
  printf '[k8s-kind] ensure ingress-nginx controller\n'
  kubectl --context "${KUBE_CONTEXT}" apply -f "${INGRESS_NGINX_MANIFEST}" >/dev/null
  if ! wait_for_job_if_exists ingress-nginx-admission-create; then
    dump_ingress_debug
    return 1
  fi
  if ! wait_for_job_if_exists ingress-nginx-admission-patch; then
    dump_ingress_debug
    return 1
  fi
  if ! kubectl --context "${KUBE_CONTEXT}" -n ingress-nginx wait \
    --for=condition=Available deployment/ingress-nginx-controller \
    --timeout=300s; then
    dump_ingress_debug
    return 1
  fi
  if ! wait_for_secret ingress-nginx-admission; then
    dump_ingress_debug
    return 1
  fi
  if ! wait_for_admission_endpoint; then
    dump_ingress_debug
    return 1
  fi
}

apply_overlay() {
  local attempt

  for attempt in $(seq 1 10); do
    if kubectl --context "${KUBE_CONTEXT}" apply -k "${KUSTOMIZE_DIR}"; then
      return 0
    fi

    printf '[k8s-kind] overlay apply failed, retrying (%s/10)\n' "${attempt}"
    sleep 5
  done

  printf 'failed to apply local kustomize overlay after retries\n' >&2
  return 1
}

require docker
require kind
require kubectl

if ! kind get clusters | grep -qx "${CLUSTER_NAME}"; then
  printf '[k8s-kind] create cluster %s\n' "${CLUSTER_NAME}"
  kind_create_cluster
else
  printf '[k8s-kind] reuse cluster %s\n' "${CLUSTER_NAME}"
fi

install_ingress_nginx

"${ROOT_DIR}/deploy/k8s/build-images.sh"

load_image gateway-api
load_image user-rpc
load_image activity-rpc
load_image notification-rpc
load_image im-rpc
load_image audit-worker
load_image postgres
load_image swagger-ui
load_image prometheus
load_image grafana

printf '[k8s-kind] apply %s\n' "${KUSTOMIZE_DIR}"
apply_overlay

restart postgres
restart rabbitmq
restart redis
restart user-rpc
restart activity-rpc
restart notification-rpc
restart im-rpc
restart audit-worker
restart gateway-api
restart swagger-ui
restart pghero
restart prometheus
restart grafana

rollout postgres
rollout rabbitmq
rollout redis
rollout user-rpc
rollout activity-rpc
rollout notification-rpc
rollout im-rpc
rollout audit-worker
rollout gateway-api
rollout swagger-ui
rollout pghero
rollout prometheus
rollout grafana

kubectl --context "${KUBE_CONTEXT}" -n "${NAMESPACE}" get pods,svc,ingress

cat <<EOF

local ingress entrypoints:
  api:        http://api.localhost:8080/
  swagger:    http://api.localhost:8080/docs/
  pghero:     http://pghero.localhost:8080/
  prometheus: http://prometheus.localhost:8080/
  grafana:    http://grafana.localhost:8080/
  rabbitmq:   http://rabbitmq.localhost:8080/

direct infra ports:
  postgres: localhost:5432
  redis:    localhost:6379
  rabbitmq: localhost:5672
  rabbitmq ui: http://localhost:15672/
EOF
