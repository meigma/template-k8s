# Technical Notes

- The template's current example operator is `NginxDeployment` in `example.meigma.io/v1alpha1`; it owns a same-named ConfigMap, Deployment, and ClusterIP Service.
- Because owned children are same-named, `NginxDeployment` names are validated as Service-safe DNS labels no longer than 63 characters. Inline `spec.config` is capped at 64 KiB before it is copied into a ConfigMap.
- The demo workload uses `nginxinc/nginx-unprivileged:stable` on port 8080 with Restricted-compatible pod/container security settings and resource requests.
- Moon is the task front door for this template. Use `moon run root:generate`, `moon run root:manifests`, `moon run root:test`, `moon run root:lint`, `moon ci --summary minimal`, and `moon run root:test-e2e` for controller/API changes. `root:test` wraps `setup-envtest` so agents do not need plain `go test ./...` to find envtest binaries.
- Repo-local Kubernetes operator lessons live in `.agents/skills/k8s-operator/SKILL.md`.
