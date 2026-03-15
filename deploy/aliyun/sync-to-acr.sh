#!/usr/bin/env bash

set -euo pipefail

COMPOSE_FILE="docker-compose.yml"
OUTPUT_FILE="docker-compose.aliyun.yml"
ENV_FILE="deploy/aliyun/compose.env.example"
DO_GENERATE=1
DO_SYNC=1
APP_PLATFORMS="${APP_PLATFORMS:-linux/amd64}"
MIRROR_PLATFORMS="${MIRROR_PLATFORMS:-linux/amd64}"

log() {
  printf '[sync-to-acr] %s\n' "$*"
}

die() {
  printf '[sync-to-acr] %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage:
  ./deploy/aliyun/sync-to-acr.sh [compose-file] [output-file] [options]

Options:
  --compose-file FILE    Source compose file. Default: docker-compose.yml
  --output-file FILE     Generated aliyun compose file. Default: docker-compose.aliyun.yml
  --env-file FILE        Env file with ACR settings. Default: deploy/aliyun/compose.env.example
  --generate-only        Rewrite the aliyun compose file but do not push images
  --sync-only            Push images to ACR but do not rewrite the aliyun compose file
  --skip-sync            Alias of --generate-only
  --skip-generate        Alias of --sync-only
  --help                 Show this message

Environment variables:
  ALIYUN_REGISTRY
  ALIYUN_APP_NAMESPACE
  ALIYUN_MIRROR_NAMESPACE
  APP_TAG
EOF
}

read_env_file_value() {
  local key="$1"

  [[ -f "$ENV_FILE" ]] || return 0

  awk -F= -v wanted="$key" '
    /^[[:space:]]*#/ { next }
    NF < 2 { next }
    $1 == wanted {
      value = substr($0, index($0, "=") + 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      gsub(/^["'"'"']|["'"'"']$/, "", value)
      print value
      exit
    }
  ' "$ENV_FILE"
}

value_or_default() {
  local key="$1"
  local fallback="$2"
  local current="${!key:-}"
  local from_file=""

  if [[ -n "$current" ]]; then
    printf '%s' "$current"
    return 0
  fi

  from_file="$(read_env_file_value "$key")"
  if [[ -n "$from_file" ]]; then
    printf '%s' "$from_file"
    return 0
  fi

  printf '%s' "$fallback"
}

image_repo_and_tag() {
  local ref="$1"
  local without_digest="${ref%@*}"
  local last_segment="${without_digest##*/}"
  local repo="$last_segment"
  local tag="latest"

  if [[ "$last_segment" == *:* ]]; then
    repo="${last_segment%:*}"
    tag="${last_segment##*:}"
  fi

  printf '%s\t%s\n' "$repo" "$tag"
}

build_target_template() {
  local service="$1"
  printf '${ALIYUN_REGISTRY:-%s}/${ALIYUN_APP_NAMESPACE:-%s}/%s:${APP_TAG:-%s}' \
    "$ALIYUN_REGISTRY" "$ALIYUN_APP_NAMESPACE" "$service" "$APP_TAG"
}

build_target_resolved() {
  local service="$1"
  printf '%s/%s/%s:%s' \
    "$ALIYUN_REGISTRY" "$ALIYUN_APP_NAMESPACE" "$service" "$APP_TAG"
}

mirror_target_template() {
  local repo="$1"
  local tag="$2"
  printf '${ALIYUN_REGISTRY:-%s}/${ALIYUN_MIRROR_NAMESPACE:-%s}/%s:%s' \
    "$ALIYUN_REGISTRY" "$ALIYUN_MIRROR_NAMESPACE" "$repo" "$tag"
}

mirror_target_resolved() {
  local repo="$1"
  local tag="$2"
  printf '%s/%s/%s:%s' \
    "$ALIYUN_REGISTRY" "$ALIYUN_MIRROR_NAMESPACE" "$repo" "$tag"
}

ensure_commands() {
  command -v docker >/dev/null 2>&1 || die "docker is required"
  command -v jq >/dev/null 2>&1 || die "jq is required"
  docker buildx version >/dev/null 2>&1 || die "docker buildx is required"
}

classify_registry() {
  case "$ALIYUN_REGISTRY" in
    *.personal.cr.aliyuncs.com)
      printf '%s' "personal-new"
      ;;
    *-registry.cn-*.cr.aliyuncs.com)
      printf '%s' "enterprise"
      ;;
    registry.cn-*.aliyuncs.com)
      printf '%s' "personal-legacy"
      ;;
    *)
      printf '%s' "unknown"
      ;;
  esac
}

docker_config_has_registry_hint() {
  local config_file="${HOME}/.docker/config.json"

  [[ -f "$config_file" ]] || return 1

  jq -e --arg registry "$ALIYUN_REGISTRY" '
    (.auths[$registry] != null) or
    (.credHelpers[$registry] != null) or
    (.credsStore != null)
  ' "$config_file" >/dev/null 2>&1
}

print_preflight() {
  local registry_kind=""

  registry_kind="$(classify_registry)"
  case "$registry_kind" in
    personal-new)
      log "registry endpoint looks like a new personal ACR instance: ${ALIYUN_REGISTRY}"
      ;;
    personal-legacy)
      log "registry endpoint looks like a legacy personal ACR instance: ${ALIYUN_REGISTRY}"
      ;;
    enterprise)
      log "registry endpoint looks like an enterprise ACR instance: ${ALIYUN_REGISTRY}"
      ;;
    *)
      log "warning: unrecognized ACR registry endpoint: ${ALIYUN_REGISTRY}"
      log "warning: use the exact login server shown in the ACR console"
      ;;
  esac

  if [[ "$ALIYUN_REGISTRY" == "registry.cn-hangzhou.aliyuncs.com" ]]; then
    log "warning: the default registry is only correct for old Hangzhou personal instances"
    log "warning: new personal instances use crpi-*.personal.cr.aliyuncs.com"
    log "warning: enterprise instances use <instance>-registry.cn-hangzhou.cr.aliyuncs.com"
  fi

  if docker_config_has_registry_hint; then
    log "docker config contains credentials or a credential helper for ${ALIYUN_REGISTRY}"
  else
    log "warning: no local docker credential hint found for ${ALIYUN_REGISTRY}"
    log "warning: run docker login --username=<your-acr-login-name> ${ALIYUN_REGISTRY}"
  fi

  log "required namespaces: ${ALIYUN_APP_NAMESPACE}, ${ALIYUN_MIRROR_NAMESPACE}"
  log "if ACR automatic repository creation is disabled, create all target repositories first"
  log "app image platforms: ${APP_PLATFORMS}"
  if [[ -n "$MIRROR_PLATFORMS" ]]; then
    log "third-party mirror platforms: ${MIRROR_PLATFORMS}"
  else
    log "third-party mirror platforms: all upstream platforms"
  fi
}

parse_args() {
  local positional=0

  while (($# > 0)); do
    case "$1" in
      --compose-file)
        [[ $# -ge 2 ]] || die "--compose-file requires a value"
        COMPOSE_FILE="$2"
        shift 2
        ;;
      --output-file)
        [[ $# -ge 2 ]] || die "--output-file requires a value"
        OUTPUT_FILE="$2"
        shift 2
        ;;
      --env-file)
        [[ $# -ge 2 ]] || die "--env-file requires a value"
        ENV_FILE="$2"
        shift 2
        ;;
      --generate-only|--skip-sync)
        DO_SYNC=0
        shift
        ;;
      --sync-only|--skip-generate)
        DO_GENERATE=0
        shift
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      --*)
        die "unknown option: $1"
        ;;
      *)
        positional=$((positional + 1))
        if [[ "$positional" -eq 1 ]]; then
          COMPOSE_FILE="$1"
        elif [[ "$positional" -eq 2 ]]; then
          OUTPUT_FILE="$1"
        else
          die "too many positional arguments"
        fi
        shift
        ;;
    esac
  done
}

write_header() {
  cat <<EOF
# Generated by ./deploy/aliyun/sync-to-acr.sh from ${COMPOSE_FILE}.
# Start with:
#   docker compose --env-file ${ENV_FILE} -f ${OUTPUT_FILE} up -d

EOF
}

generate_compose() {
  local service_targets="$1"
  local output_tmp="$2"

  {
    write_header
    awk '
      function indent_width(s,    m) {
        match(s, /^ */)
        return RLENGTH
      }

      FNR == NR {
        service = $1
        target[service] = $2
        kind[service] = $4
        next
      }

      {
        line = $0

        if (skip_build) {
          indent = indent_width(line)
          if (line ~ /^[[:space:]]*$/ || indent > 4) {
            next
          }
          skip_build = 0
        }

        if (line ~ /^services:[[:space:]]*$/) {
          print line
          next
        }

        if (line ~ /^  [A-Za-z0-9_.-]+:[[:space:]]*$/) {
          current_service = substr(line, 3)
          sub(/:[[:space:]]*$/, "", current_service)
          print line
          next
        }

        if (current_service != "" && kind[current_service] == "build" && line ~ /^    build:[[:space:]]*$/) {
          print "    image: " target[current_service]
          skip_build = 1
          next
        }

        if (current_service != "" && line ~ /^    image:[[:space:]]*/) {
          print "    image: " target[current_service]
          next
        }

        if (current_service != "" && indent_width(line) <= 2 && line !~ /^[[:space:]]*$/) {
          current_service = ""
        }

        print line
      }
    ' "$service_targets" "$COMPOSE_FILE"
  } >"$output_tmp"

  mv "$output_tmp" "$OUTPUT_FILE"
  log "generated ${OUTPUT_FILE}"
}

sync_mirror_images() {
  local mirror_pairs="$1"
  local source_image=""
  local target_image=""

  [[ -s "$mirror_pairs" ]] || {
    log "no third-party images to sync"
    return 0
  }

  log "syncing third-party images to ${ALIYUN_REGISTRY}/${ALIYUN_MIRROR_NAMESPACE}"
  while IFS=$'\t' read -r source_image target_image; do
    [[ -n "$source_image" ]] || continue
    log "copy ${source_image} -> ${target_image}"
    if [[ -n "$MIRROR_PLATFORMS" ]]; then
      docker buildx imagetools create --platform "$MIRROR_PLATFORMS" --tag "$target_image" "$source_image"
    else
      docker buildx imagetools create --tag "$target_image" "$source_image"
    fi
  done <"$mirror_pairs"
}

resolve_dockerfile_path() {
  local context="$1"
  local dockerfile="$2"

  if [[ "$dockerfile" == /* ]]; then
    printf '%s' "$dockerfile"
    return 0
  fi

  dockerfile="${dockerfile#./}"
  printf '%s/%s' "$context" "$dockerfile"
}

sync_app_images() {
  local service_targets="$1"
  local service=""
  local target_image=""
  local kind=""
  local build_context=""
  local build_dockerfile=""
  local resolved_dockerfile=""

  log "building and pushing app images to ${ALIYUN_REGISTRY}/${ALIYUN_APP_NAMESPACE}"
  while IFS=$'\t' read -r service _ target_image kind _ build_context build_dockerfile; do
    [[ "$kind" == "build" ]] || continue
    resolved_dockerfile="$(resolve_dockerfile_path "$build_context" "$build_dockerfile")"

    log "buildx ${service} -> ${target_image}"
    docker buildx build \
      --platform "$APP_PLATFORMS" \
      --pull \
      --push \
      --tag "$target_image" \
      --file "$resolved_dockerfile" \
      "$build_context"
  done <"$service_targets"
}

main() {
  local tmp_dir=""
  local compose_json=""
  local service_targets=""
  local mirror_pairs=""
  local output_tmp=""
  local build_services_count=""
  local third_party_count=""

  parse_args "$@"
  ensure_commands

  [[ -f "$COMPOSE_FILE" ]] || die "compose file not found: ${COMPOSE_FILE}"

  ALIYUN_REGISTRY="$(value_or_default ALIYUN_REGISTRY "registry.cn-hangzhou.aliyuncs.com")"
  ALIYUN_APP_NAMESPACE="$(value_or_default ALIYUN_APP_NAMESPACE "pallink")"
  ALIYUN_MIRROR_NAMESPACE="$(value_or_default ALIYUN_MIRROR_NAMESPACE "pallink-mirror")"
  APP_TAG="$(value_or_default APP_TAG "latest")"
  APP_PLATFORMS="$(value_or_default APP_PLATFORMS "$APP_PLATFORMS")"
  MIRROR_PLATFORMS="$(value_or_default MIRROR_PLATFORMS "$MIRROR_PLATFORMS")"

  tmp_dir="$(mktemp -d)"
  trap "rm -rf \"$tmp_dir\"" EXIT

  compose_json="${tmp_dir}/compose.json"
  service_targets="${tmp_dir}/service-targets.tsv"
  mirror_pairs="${tmp_dir}/mirror-pairs.tsv"
  output_tmp="${tmp_dir}/aliyun-compose.yml"

  docker compose -f "$COMPOSE_FILE" config --format json >"$compose_json"

  : >"$service_targets"
  : >"$mirror_pairs"

  jq -r '.services | to_entries[] | [.key, (.value.build != null), (.value.image // "-"), (.value.build.context // "-"), (.value.build.dockerfile // "-")] | @tsv' "$compose_json" |
    while IFS=$'\t' read -r service has_build source_image build_context build_dockerfile; do
      local repo=""
      local tag=""
      local target_template=""
      local target_resolved=""

      if [[ "$has_build" == "true" ]]; then
        target_template="$(build_target_template "$service")"
        target_resolved="$(build_target_resolved "$service")"
        printf '%s\t%s\t%s\tbuild\t%s\t%s\t%s\n' \
          "$service" "$target_template" "$target_resolved" "-" "$build_context" "$build_dockerfile" >>"$service_targets"
        continue
      fi

      [[ -n "$source_image" && "$source_image" != "-" ]] || continue

      IFS=$'\t' read -r repo tag <<EOF
$(image_repo_and_tag "$source_image")
EOF

      target_template="$(mirror_target_template "$repo" "$tag")"
      target_resolved="$(mirror_target_resolved "$repo" "$tag")"

      printf '%s\t%s\t%s\tmirror\t%s\t%s\t%s\n' \
        "$service" "$target_template" "$target_resolved" "$source_image" "-" "-" >>"$service_targets"
      printf '%s\t%s\n' "$source_image" "$target_resolved" >>"$mirror_pairs"
    done

  sort -u "$mirror_pairs" -o "$mirror_pairs"
  awk -F '\t' '
    {
      if (seen[$2] && seen[$2] != $1) {
        printf "[sync-to-acr] target image collision: %s maps to both %s and %s\n", $2, seen[$2], $1 > "/dev/stderr"
        exit 1
      }
      seen[$2] = $1
    }
  ' "$mirror_pairs"

  build_services_count="$(awk -F '\t' '$4 == "build" { count++ } END { print count + 0 }' "$service_targets")"
  third_party_count="$(wc -l <"$mirror_pairs" | tr -d ' ')"
  log "resolved ${build_services_count} app services and ${third_party_count} third-party images"
  print_preflight

  if [[ "$DO_GENERATE" -eq 1 ]]; then
    generate_compose "$service_targets" "$output_tmp"
  fi

  if [[ "$DO_SYNC" -eq 1 ]]; then
    sync_mirror_images "$mirror_pairs"
    sync_app_images "$service_targets"
  fi

  log "done"
}

main "$@"
