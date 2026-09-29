# Benchmarks unit converter

A small, self-contained Go web app for converting common metric and imperial length measurements. Enter a value in any row and the remaining rows update automatically. Fractional inches can be rounded to standard shop increments from 1/2 through 1/64. The displayed units can be customized per browser; millimeters, fractional inches, and decimal inches are shown by default.

This app is one component of the `unit-conversions` repository. The commands
below are run from this `mm-inches` directory unless noted otherwise.

## Run locally

```sh
go run .
```

Open <http://localhost:8080>. Set `PORT` to listen on another port.

## Test

```sh
go test ./...
```

## Build the container

```sh
docker build -t tyler180/mm-inches:local .
./scripts/smoke-test.sh tyler180/mm-inches:local
```

The final image is distroless, runs as a non-root user, and writes no application data to disk.

## CI/CD

Pull requests run formatting, race-enabled tests, `go vet`, a binary build,
Kustomize rendering, and a hardened container smoke test.

After changes are merged to `main`, create and push a semantic version tag:

```sh
git switch main
git pull --ff-only
git tag -a mm-inches-v0.1.0 -m "Release mm-inches v0.1.0"
git push origin mm-inches-v0.1.0
```

The release workflow publishes `tyler180/mm-inches:v0.1.0` for `linux/amd64`
with provenance and an SBOM. It then opens a pull request in
`tyler180/talos-gitops` using an image reference such as
`tyler180/mm-inches:v0.1.0@sha256:...`. This keeps the version readable while
pinning the exact image content. Merge and sync that GitOps pull request when
the release is ready for the cluster.

Configure these Actions secrets in this repository before merging the CI/CD
setup:

- `DOCKERHUB_TOKEN`: a Docker Hub access token that can push `tyler180/mm-inches`.
- `GITOPS_TOKEN`: a fine-grained GitHub token with Contents read/write and Pull
  requests read/write access to `tyler180/talos-gitops`.

## Deploy to Kubernetes

1. Replace `ghcr.io/REPLACE_ME/mm-inches:latest` in `deploy/deployment.yaml` with the image you pushed. Prefer an immutable version tag or digest for GitOps.
2. Validate the rendered resources:

   ```sh
   kubectl kustomize deploy
   kubectl apply --dry-run=server -k deploy
   ```

3. Add `deploy/` to the appropriate Argo CD application, or apply it manually:

   ```sh
   kubectl apply -k deploy
   kubectl -n unit-conversions rollout status deployment/mm-inches
   ```

The included Service is cluster-internal. Expose it with the Gateway or Ingress pattern already used by your cluster rather than assuming a specific controller or certificate setup.

## Endpoints

- `/` — converter interface
- `/api/convert` — conversion endpoint used by the interface
- `/healthz` — liveness and readiness check
