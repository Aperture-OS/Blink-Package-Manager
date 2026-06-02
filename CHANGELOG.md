<p align="center">
<br/>
<img src="https://github.com/Aperture-OS/branding/blob/main/Logo-Bright/logo-bright.png?raw=true" alt="ApertureOS Logo" width=100>
<h3 align="center">Changelog</h3>
<p align="center">
All notable changes to Blink Package Manager will be documented in this file.</p>
</p>
<p align="center">
<img alt="Static Badge" src="https://img.shields.io/badge/Blink_Package_Manager-v0.2.0-6d7592?logo=github&labelColor=45455e&link=https%3A%2F%2Fgithub.com%2FAperture-OS%2Fblink-package-manager%2Freleases">
</p>

## [v0.2.0-alpha] - 2026-03-09

### Added
- Unit tests for all core modules: structs, utils, config, manifest, globals, source, repository, lock.
- GitHub Actions CI workflow with lint, security scanning, build, and test jobs.
- `SECURITY.md` with vulnerability reporting guidelines and security practices.
- `CHANGELOG.md` for tracking project changes.
- `docs/` folder with CLI usage, manifest format, repository workflow, and packaging documentation.
- Custom HTTP client with HTTPS-only redirect enforcement and redirect limit.
- Download size limit (2 GB) to prevent denial-of-service via large files.
- URL scheme validation (HTTPS-only) before source downloads.
- `PackageInfo.Validate()` call in install and uninstall flows.
- Release workflow for building Linux artifacts (amd64 + arm64).

### Fixed
- `gofmt` formatting issues in `structs.go`.
- `checkDirAndCreate` used overly permissive `os.ModePerm` (0777), now uses `0750`.
- `saveManifest` did not sync and close file before rename, risking data loss.
- `IsLocked` created lock file with 0644 permissions, now uses 0600.
- Downloaded source files created with world-readable permissions, now use 0640.

### Security
- Replaced bare `http.Get()` with a hardened HTTP client that blocks non-HTTPS redirects.
- Added `io.LimitReader` to cap download size at 2 GB.
- Added source URL validation requiring HTTPS scheme before downloads.
- Added package recipe validation (`Validate()`) before install/uninstall to enforce allowlisted env keys, SHA256 format, and source types.

## [v0.1.0-alpha] - 2025-12-27

### Added
- Initial release of Blink Package Manager.
- Source-based package installation (`toCompile`) and precompiled binary installation (`preCompiled`).
- Dependency resolution with topological sorting (DFS).
- Optional dependency selection with interactive prompts.
- Repository sync via Git clone/fetch with GPG signature verification.
- TOML-based configuration and manifest management.
- File-based locking to prevent concurrent operations.
- CLI commands: `install`, `uninstall`, `get`, `search`, `sync`, `update`, `clean`, `version`, `support`, `completion`.
- SHA256 source integrity verification.
- Path traversal protection for precompiled packages.

**&copy; Copyright Aperture OS 2025-2026**
**All Rights Reserved!**
