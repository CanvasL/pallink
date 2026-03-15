#!/usr/bin/env bash

set -euo pipefail

FILE="gateway/swagger.json"
SWAGGER_API_HOST="${SWAGGER_API_HOST:-}"
SWAGGER_API_BASE_PATH="${SWAGGER_API_BASE_PATH:-/}"
SWAGGER_API_SCHEMES="${SWAGGER_API_SCHEMES:-}"

usage() {
  cat <<'EOF'
Usage:
  ./deploy/swagger/patch.sh [--file FILE]
  ./deploy/swagger/patch.sh FILE
EOF
}

die() {
  printf 'swagger patch failed: %s\n' "$*" >&2
  exit 1
}

while (($# > 0)); do
  case "$1" in
    --file|-file)
      [[ $# -ge 2 ]] || die "--file requires a value"
      FILE="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      [[ $# -eq 1 ]] || die "unexpected argument: $1"
      FILE="$1"
      shift
      ;;
  esac
done

command -v jq >/dev/null 2>&1 || die "jq is required"
[[ -f "$FILE" ]] || die "file not found: $FILE"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

jq --indent 2 \
  --arg swagger_api_host "$SWAGGER_API_HOST" \
  --arg swagger_api_base_path "$SWAGGER_API_BASE_PATH" \
  --arg swagger_api_schemes "$SWAGGER_API_SCHEMES" '
  def route_tag($route):
    if ($route | startswith("/user/")) then "user"
    elif ($route | startswith("/activity/")) then "activity"
    elif ($route | startswith("/notification/")) then "notification"
    elif ($route | startswith("/im/")) then "im"
    else ""
    end;

  def route_requires_auth($route):
    if $route == "/user/login" or $route == "/user/register" then false
    elif ($route | startswith("/activity/public/")) then false
    else true
    end;

  def is_http_method($method):
    ["get", "post", "put", "delete", "patch", "head", "options"]
    | index($method | ascii_downcase) != null;

  .tags = [
    {"name": "user", "description": "用户"},
    {"name": "activity", "description": "活动"},
    {"name": "notification", "description": "通知"},
    {"name": "im", "description": "私聊"}
  ]
  | if $swagger_api_host != "" then
      .host = $swagger_api_host
    else
      del(.host)
    end
  | .basePath = $swagger_api_base_path
  | if $swagger_api_schemes != "" then
      .schemes = ($swagger_api_schemes | split(",") | map(gsub("^ +| +$"; "")) | map(select(. != "")))
    else
      del(.schemes)
    end
  | .securityDefinitions = {
      "BearerAuth": {
        "type": "apiKey",
        "name": "Authorization",
        "in": "header",
        "description": "JWT Bearer token. Example: Authorization: Bearer <token>"
      }
    }
  | if (.paths | type) != "object" then
      error("paths field missing or invalid")
    else
      .
    end
  | .paths |= with_entries(
      .key as $route
      | (route_tag($route)) as $tag
      | if $tag == "" or (.value | type) != "object" then
          .
        else
          .value |= with_entries(
            if is_http_method(.key) and (.value | type) == "object" then
              .value |= (
                .tags = [$tag]
                | del(.schemes)
                | if route_requires_auth($route) then
                    .security = [{"BearerAuth": []}]
                  else
                    del(.security)
                  end
              )
            else
              .
            end
          )
        end
    )
' "$FILE" >"$TMP_FILE"

mv "$TMP_FILE" "$FILE"
