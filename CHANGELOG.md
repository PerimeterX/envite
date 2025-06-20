# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.11](https://github.com/PerimeterX/envite/compare/v0.0.10...v0.0.11)

### Added

- Added docker runtime awareness with support for the following:
  - Docker Desktop
  - Colima (with 3-second network latency)
  - Podman
  - Rancher Desktop
  - Lima
  - OrbStack
  - Minikube
  - ContainerD
  - Finch
- `ExtractRuntimeInfo` function to detect runtime type from Docker client info
- Runtime-specific internal hostname mapping (e.g., `host.docker.internal`, `host.lima.internal`)
- Network latency configuration for runtimes that require startup delays
- Improved error handling with error wrapping to provide better error messages

## [0.0.10](https://github.com/PerimeterX/envite/compare/v0.0.9...v0.0.10)

### Fixed

- Cleanup phase still failed sometimes after previous fix with different message, now covering both.

## [0.0.9](https://github.com/PerimeterX/envite/compare/v0.0.8...v0.0.9)

### Fixed

- Cleanup phase failed when using the same image for multiple components due to a failure in removing the image  
more than once. To fix this - ignored that specific issue "reference does not exist" when that error is returned   
from the docker remove request

## [0.0.8](https://github.com/PerimeterX/envite/compare/v0.0.7...v0.0.8)

### Added

- Postgres Seed functionality.

## [0.0.7](https://github.com/PerimeterX/envite/compare/v0.0.6...v0.0.7)

### Fixed

- upgrade docker(moby) lib and update usages to deprecated structs see https://github.com/moby/moby/releases/tag/v27.0.1
- upgraded go version to v1.22

## [0.0.6](https://github.com/PerimeterX/envite/compare/v0.0.5...v0.0.6)

### Added

- Redis Seed functionality.

## [0.0.5](https://github.com/PerimeterX/envite/compare/v0.0.4...v0.0.5) - 2024-02-25

### Fixed

- Fix error parsing null configs.

## [0.0.4](https://github.com/PerimeterX/envite/compare/v0.0.3...v0.0.4) - 2024-02-21

### Fixed

- Fix all dependency vulnerabilities.

## [0.0.3](https://github.com/PerimeterX/envite/compare/v0.0.2...v0.0.3) - 2024-02-17

Prepare for open source.

### Added

- CLI support.
- Go releaser support.
- Unit tests.
- Go docs.
- Improved README.md.
- Added several small changes to the SDK API to allow smoother experience in the CLI.
