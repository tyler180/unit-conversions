#!/usr/bin/env bash

set -euo pipefail

image="${1:-mm-inches:ci}"
container="mm-inches-smoke-${RANDOM}"

cleanup() {
  docker rm -f "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT

configured_user="$(docker image inspect --format '{{.Config.User}}' "$image")"
if [[ "$configured_user" != "65532:65532" ]]; then
  echo "expected image user 65532:65532, got ${configured_user:-<empty>}" >&2
  exit 1
fi

docker run --detach \
  --name "$container" \
  --read-only \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --tmpfs /tmp:rw,noexec,nosuid,size=16m \
  --publish 127.0.0.1::8080 \
  "$image" >/dev/null

host_port="$(
  docker inspect \
    --format '{{(index (index .NetworkSettings.Ports "8080/tcp") 0).HostPort}}' \
    "$container"
)"
base_url="http://127.0.0.1:${host_port}"

for _ in {1..30}; do
  if curl --fail --silent --show-error "${base_url}/healthz" | grep -qx 'ok'; then
    break
  fi
  sleep 1
done

curl --fail --silent --show-error "${base_url}/healthz" | grep -qx 'ok'

response="$(
  curl --fail --silent --show-error \
    --header 'Content-Type: application/json' \
    --data '{"unit":"millimeters","value":"426","decimalPlaces":4,"fractionDenominator":64}' \
    "${base_url}/api/convert"
)"

if [[ "$response" != *'"fractional-inches":"16 49/64"'* ]]; then
  echo "unexpected conversion response: $response" >&2
  exit 1
fi

echo "container smoke test passed"
