#!/usr/bin/env sh

set -eu

INPUT_FILE="${1:-}"
OUTPUT_FILE="${2:-}"

if [ -z "$INPUT_FILE" ] || [ -z "$OUTPUT_FILE" ]; then
  echo "usage: ./deploy/swagger/render.sh <input-file> <output-file>" >&2
  exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
  echo "swagger render failed: file not found: $INPUT_FILE" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"

node <<'EOF' "$INPUT_FILE" "$OUTPUT_FILE"
const fs = require("fs");

const input = process.argv[1];
const output = process.argv[2];
const host = process.env.SWAGGER_API_HOST || "";
const basePath = process.env.SWAGGER_API_BASE_PATH || "/";
const schemes = (process.env.SWAGGER_API_SCHEMES || "")
  .split(",")
  .map((item) => item.trim())
  .filter(Boolean);

const doc = JSON.parse(fs.readFileSync(input, "utf8"));

if (host) {
  doc.host = host;
} else {
  delete doc.host;
}

doc.basePath = basePath;

if (schemes.length > 0) {
  doc.schemes = schemes;
} else {
  delete doc.schemes;
}

fs.writeFileSync(output, JSON.stringify(doc, null, 2) + "\n");
EOF
