# Unit conversions

A monorepo for small, self-hosted measurement conversion tools.

## Converters

| Directory | Measurement | Status |
| --- | --- | --- |
| [`mm-inches/`](mm-inches/) | Length | Available |
| [`volume/`](volume/) | Volume | Planned |
| [`weight/`](weight/) | Weight | Planned |

Each converter is kept in its own directory with its own application code,
tests, container definition, deployment manifests, and documentation. Shared
repository automation lives in `.github/` and is scoped to the converter it
builds.

Release tags are namespaced by converter so their versions can evolve
independently. For example, the length converter uses tags such as
`mm-inches-v0.1.3`.
