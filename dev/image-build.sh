#!/usr/bin/env bash
# Build the operator image for the host architecture with melange + apko and load
# it into Docker. Mirrors the release path (melange.yaml + apko.yaml) so local e2e
# exercises the same Wolfi nonroot image that ships. melange needs Linux, so
# `--runner docker` runs the build in a Linux container (Docker must be running).
#
# Override the loaded tag with IMG (default template-k8s:dev).
set -euo pipefail

image="${IMG:-template-k8s:dev}"
arch="$(go env GOARCH)"

rm -rf packages image.tar
melange keygen melange.rsa
melange build melange.yaml \
  --arch "$arch" \
  --runner docker \
  --signing-key melange.rsa \
  --source-dir .
# apko single-arch build loads as <tag>-<arch>; retag to the requested tag.
apko build apko.yaml "$image" image.tar \
  --arch "$arch" \
  --keyring-append ./melange.rsa.pub
docker load < image.tar
docker tag "${image}-${arch}" "$image"
echo "loaded ${image} (host arch ${arch})"
