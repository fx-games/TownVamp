#!/bin/sh

set -eu

cd "$(dirname "$0")"

mkdir -p ./builds/dev

go build -trimpath -o ./builds/dev/LuxAeterna -tags "ebitensinglethread development" .

./builds/dev/LuxAeterna "$@"
