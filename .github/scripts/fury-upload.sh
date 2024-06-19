#!/usr/bin/env bash
# Originally adapated from:
# https://github.com/goreleaser/goreleaser/blob/main/scripts/fury-upload.sh
set -euo pipefail

# FURY_PUSH_TOKEN is a secret token that is used to push packages to
# fury (https://fury.io). It is a required environment variable.
FURY_PUSH_TOKEN="${FURY_PUSH_TOKEN:-}"

# allowed_exts is an array of allowed file extensions that can be
# uploaded to fury.
allowed_exts=("deb" "rpm" "apk")

# Get the file extension
file_ext="${1##*.}"

is_allowed=false
for ext in "${allowed_exts[@]}"; do
  if [[ "$ext" == "$file_ext" ]]; then
    is_allowed=true
    break
  fi
done

if [[ "$is_allowed" == "false" ]]; then
  exit 0
fi

cd dist
echo "uploading $1"
status="$(curl -s -q -o /dev/null -w "%{http_code}" -F package="@$1" "https://${FURY_PUSH_TOKEN}@push.fury.io/rgst-io/")"
echo "got: $status"
if [[ "$status" == "200" ]] || [[ "$status" == "409" ]]; then
  exit 0
fi

# Otherwise, exit with an error
exit 1
