# Security Audit Report — Blink Package Manager

**Date:** 2025  
**Scope:** All Go source files in `src/`  
**Standard:** OWASP Top 10, CWE/SANS Top 25

---

## Executive Summary

Eight security vulnerabilities were identified during the initial audit. All have been remediated. The fixes harden network I/O, input validation, file permissions, and atomic writes. No critical issues remain open.

---

## Findings

### 1. Unrestricted HTTP Redirect Following

| Field | Value |
|-------|-------|
| **Severity** | High |
| **OWASP** | A10 — Server-Side Request Forgery |
| **File** | `src/source.go` lines 40–50 |
| **Before** | Default `http.Client` followed arbitrary redirects, including `http://` downgrades. |
| **Fix** | Custom `httpClient` with `CheckRedirect` that rejects non-HTTPS targets and caps redirect depth at 10. |
| **Verified** | Build + vet pass; unit test in `source_test.go`. |

### 2. Unbounded Download Size

| Field | Value |
|-------|-------|
| **Severity** | High |
| **OWASP** | A05 — Security Misconfiguration |
| **File** | `src/source.go` line 38, lines 87–88 |
| **Before** | Response body was copied without any size limit; a malicious server could exhaust disk space. |
| **Fix** | Added `maxDownloadSize = 2 GB` constant and wrapped `resp.Body` in `io.LimitReader`. |
| **Verified** | Build + vet pass. |

### 3. Non-HTTPS Source URLs Accepted

| Field | Value |
|-------|-------|
| **Severity** | High |
| **OWASP** | A02 — Cryptographic Failures |
| **File** | `src/source.go` lines 73–78 |
| **Before** | Any URL scheme was accepted, allowing plain HTTP downloads. |
| **Fix** | Parse URL with `net/url.Parse` and reject anything other than `https://`. |
| **Verified** | Build + vet pass; `Validate()` also checks scheme. |

### 4. Missing Recipe Validation Before Build

| Field | Value |
|-------|-------|
| **Severity** | High |
| **OWASP** | A03 — Injection |
| **File** | `src/package_ops.go` lines 208, 376 |
| **Before** | Package recipes were used directly without validating URL, hash, source type, or env keys. |
| **Fix** | Added `pkg.Validate()` call at the start of both `install()` and `uninstall()`. The `Validate()` method (defined in `src/structs.go` lines 85–114) checks: HTTPS-only URL, valid SHA256 hex digest, allowlisted source types, allowlisted env keys, and control character rejection in env values. |
| **Verified** | Build + vet pass; unit tests in `structs_test.go`. |

### 5. World-Writable Directory Permissions

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **OWASP** | A01 — Broken Access Control |
| **File** | `src/utils.go` line 47 |
| **Before** | `checkDirAndCreate` used mode `0777`, allowing any user to write to package directories. |
| **Fix** | Changed to `0750` (owner rwx, group r-x, others none). |
| **Verified** | Build + vet pass. |

### 6. Overly Permissive File Permissions

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **OWASP** | A01 — Broken Access Control |
| **File** | `src/source.go` line 92 |
| **Before** | Downloaded source files were created with mode `0644` (world-readable). |
| **Fix** | Changed `os.Create` to `os.OpenFile` with mode `0640` (group-readable, no world access). |
| **Verified** | Build + vet pass. |

### 7. Non-Atomic Manifest Writes

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **OWASP** | A08 — Software and Data Integrity Failures |
| **File** | `src/manifest.go` lines 81–101 |
| **Before** | Manifest was written directly in-place. A crash mid-write would corrupt the installed-package database. |
| **Fix** | Write → `Sync()` → `Close()` a temporary file, then `os.Rename` atomically over the original. File mode set to `0640`. |
| **Verified** | Build + vet pass; unit test in `manifest_test.go`. |

### 8. Lock File Created World-Readable

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **OWASP** | A01 — Broken Access Control |
| **File** | `src/lock.go` lines 44, 108 |
| **Before** | `Acquire()` and `IsLocked()` created the lock file with mode `0644`. |
| **Fix** | Both now use mode `0600` (owner-only read/write). |
| **Verified** | Build + vet pass; unit test in `lock_test.go`. |

---

## Additional Hardening Already Present

| Control | Location |
|---------|----------|
| Path traversal protection in `safeExtractToRoot` | `src/source.go` |
| SHA256 source verification | `src/source.go`, `src/utils.go` |
| GPG commit signature verification | `src/repository.go` |
| Exclusive file locking (`syscall.Flock`) | `src/lock.go` |
| Root privilege check before operations | `src/utils.go` — `requireRoot()` |

---

## Recommendations For Future Work

1. **TLS certificate pinning** — Consider pinning repository server certificates for high-value repositories.
2. **Sandboxed builds** — Run `build.prepare` / `build.install` commands inside a namespace or container to limit blast radius.
3. **Audit logging** — Write a tamper-evident log of all install/uninstall operations for forensic review.
4. **Dependency signature verification** — Extend GPG verification from repository commits to individual recipe files.

---

## Conclusion

All eight identified vulnerabilities have been remediated. The codebase now enforces HTTPS-only downloads, bounded I/O, strict input validation, least-privilege file permissions, and atomic writes. The security posture is appropriate for an alpha-stage package manager operating with root privileges.
