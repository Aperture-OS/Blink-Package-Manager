<p align="center">
<br/>
<img src="https://github.com/Aperture-OS/branding/blob/main/Logo-Bright/logo-bright.png?raw=true" alt="ApertureOS Logo" width=100>
<h3 align="center">Security Policy</h3>
<p align="center">
Guidelines for reporting security vulnerabilities in Blink Package Manager.</p>
</p>
<p align="center">
<img alt="Static Badge" src="https://img.shields.io/badge/Blink_Package_Manager-v0.2.0-6d7592?logo=github&labelColor=45455e&link=https%3A%2F%2Fgithub.com%2FAperture-OS%2Fblink-package-manager%2Freleases">
</p>

## Supported Versions

| Version        | Supported          |
| -------------- | ------------------ |
| v0.2.0-alpha   | :white_check_mark: |
| < v0.2.0-alpha | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Blink, **please do not open a public issue.**

Instead, report it privately via one of the following channels:

1. **GitHub Security Advisories:** Use the [Security tab](https://github.com/Aperture-OS/blink-package-manager/security/advisories/new) to report a vulnerability privately.
2. **Discord:** Contact a maintainer directly on our [Discord server](https://discord.com/invite/rx82u93hGD).

### What to Include

- A clear description of the vulnerability.
- Steps to reproduce.
- The affected file(s) and line(s) if known.
- Any potential impact or exploit scenario.

### Response Timeline

- **Acknowledgement:** Within 48 hours.
- **Initial assessment:** Within 7 days.
- **Fix and disclosure:** As soon as a patch is ready, coordinated with the reporter.

## Security Practices

Blink follows these security practices:

- **HTTPS-only downloads:** Source archives are only fetched over HTTPS.
- **SHA256 verification:** All downloaded sources are verified against their declared hash before extraction.
- **Allowlisted build environment:** Only a predefined set of environment variables (`CC`, `CXX`, `CFLAGS`, `CXXFLAGS`, `LDFLAGS`, `PREFIX`, `DESTDIR`, `MAKEFLAGS`, `PKG_CONFIG_PATH`) are permitted in package recipes.
- **Path traversal protection:** Extracted archives are checked for path traversal attacks before installation.
- **GPG commit verification:** Repository integrity is verified via GPG-signed commits against a trusted public key.
- **File-based locking:** Concurrent operations are prevented via exclusive file locks.
- **Download size limits:** Source downloads are capped at 2 GB to prevent denial-of-service.
- **Redirect safety:** HTTP redirects are limited and must remain on HTTPS.

## License

Blink is released under Apache 2.0, see more at [LICENSE](https://github.com/Aperture-OS/blink-package-manager/blob/main/LICENSE)

**&copy; Copyright Aperture OS 2025-2026**
**All Rights Reserved!**
