#!/usr/bin/env bash
set -euo pipefail

cluster_config="dev/ctlptl.yaml"
context="kind-template-k8s-dev"
sample_namespace="template-k8s-smoke"
sample_file="test/chainsaw/fixtures/example_v1alpha1_nginxdeployment.yaml"
smoke_port="${DEV_SMOKE_PORT:-18080}"
pf_pid=""

cleanup() {
  status="$?"
  set +e
  if [ -n "$pf_pid" ]; then
    kill "$pf_pid" >/dev/null 2>&1
    wait "$pf_pid" >/dev/null 2>&1
  fi
  kubectl --context "$context" delete -n "$sample_namespace" -f "$sample_file" --ignore-not-found=true >/dev/null 2>&1
  kubectl --context "$context" delete namespace "$sample_namespace" --ignore-not-found=true >/dev/null 2>&1
  tilt down --context "$context" --delete-namespaces >/dev/null 2>&1
  ctlptl delete -f "$cluster_config" --cascade=true --ignore-not-found >/dev/null 2>&1
  exit "$status"
}
trap cleanup EXIT

ctlptl apply -f "$cluster_config"
kubectl config use-context "$context" >/dev/null
kubectl --context "$context" wait --for=condition=Ready node --all --timeout=3m

registry_host="$(ctlptl get cluster "$context" -o template --template '{{.status.localRegistryHosting.host}}')"
if [ -z "$registry_host" ]; then
  echo "ctlptl did not report a local registry host for $context" >&2
  exit 1
fi
echo "local registry: $registry_host"

tilt ci --context "$context" --timeout 10m

kubectl --context "$context" create namespace "$sample_namespace" --dry-run=client -o yaml | kubectl --context "$context" apply -f -
kubectl --context "$context" label --overwrite namespace "$sample_namespace" pod-security.kubernetes.io/enforce=restricted
kubectl --context "$context" apply -n "$sample_namespace" -f "$sample_file"
kubectl --context "$context" wait -n "$sample_namespace" nginxdeployment/nginx-sample --for=condition=Available --timeout=3m
kubectl --context "$context" wait -n "$sample_namespace" deployment/nginx-sample --for=condition=Available --timeout=3m

kubectl --context "$context" port-forward --address localhost -n "$sample_namespace" service/nginx-sample "${smoke_port}:8080" >/tmp/template-k8s-dev-port-forward.log 2>&1 &
pf_pid="$!"

for _ in $(seq 1 30); do
  if body="$(curl -fsS "http://localhost:${smoke_port}/" 2>/dev/null)"; then
    printf '%s\n' "$body"
    test "$body" = "hello from template-k8s"
    exit 0
  fi
  sleep 1
done

cat /tmp/template-k8s-dev-port-forward.log >&2
exit 1
