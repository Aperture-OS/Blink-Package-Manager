# Packaging Workflow

This document describes how Blink builds and installs packages from source, and how precompiled packages are handled.

## Build Kinds

Blink supports two build kinds, specified by the `build.kind` field in the package recipe:

### `toCompile` — Build from Source

1. **Download:** Blink downloads the source archive from `source.url` over HTTPS.
2. **Verify:** The SHA256 hash of the downloaded archive is compared against `source.sha256`.
3. **Extract:** The archive is extracted into a build directory (`/var/blink/build/<package>/`).
4. **Environment:** Variables from `build.env` are set (only allowlisted keys: `CC`, `CXX`, `CFLAGS`, `CXXFLAGS`, `LDFLAGS`, `PREFIX`, `DESTDIR`, `MAKEFLAGS`, `PKG_CONFIG_PATH`).
5. **Prepare:** Commands in `build.prepare` are executed in order (e.g., `./configure`, patching).
6. **Install:** Commands in `build.install` are executed in order (e.g., `make install`).
7. **Record:** The package is added to the manifest.

### `preCompiled` — Install Prebuilt Binaries

1. **Download:** Blink downloads the binary archive from `source.url` over HTTPS.
2. **Verify:** SHA256 hash verification.
3. **Extract:** The archive is extracted into a staging directory.
4. **Safety check:** All extracted paths are validated for path traversal attacks (`..` or absolute paths are rejected).
5. **Copy:** Files are copied from the staging directory to the filesystem root, preserving permissions.
6. **Install:** Commands in `build.install` are executed (if any).
7. **Record:** The package is added to the manifest.

## Uninstall Flow

1. **Fetch recipe:** The package recipe is retrieved from cache or repository.
2. **Verify source:** The source archive is re-downloaded and verified.
3. **Extract:** Source is extracted to reconstruct the build environment.
4. **Environment:** Build environment variables are set.
5. **Uninstall:** Commands in `build.uninstall` are executed.
6. **Remove from manifest:** The package is removed from the installed manifest.

## Dependency Resolution

### Mandatory Dependencies

Blink resolves mandatory dependencies using a depth-first graph traversal with topological sorting:

1. The dependency graph is built recursively by fetching each dependency's recipe.
2. Cycle detection is performed using DFS.
3. Dependencies are sorted in topological order (dependencies before dependents).
4. Missing dependencies are presented to the user for confirmation.
5. Each dependency is installed in order before the requested package.

### Optional Dependencies

Optional dependencies are grouped by ID. For each group:

1. Already-installed options are identified.
2. Remaining options are presented to the user as a numbered list.
3. The user selects an option (or skips with `0`).
4. The selected option and its own dependencies are resolved and installed.

## Directory Layout

All Blink data is stored under `/var/blink/` (or `<root>/var/blink/` with `--root`):

```
/var/blink/
├── etc/
│   ├── config.toml      # Repository configuration
│   ├── manifest.toml    # Installed packages manifest
│   └── blink.lock       # Process lock file
├── repositories/        # Cloned Git repositories
│   └── <reponame>/
│       └── recipes/
├── recipes/             # Cached recipe files
├── sources/             # Downloaded source archives
└── build/               # Build and extraction directories
    └── <package>/
```

## Security Checks During Build

| Check | When | What |
|-------|------|------|
| HTTPS-only URL | Before download | Source URL must use `https://` scheme |
| SHA256 verification | After download | Archive hash must match recipe declaration |
| Env key allowlist | Before build | Only permitted environment variable names accepted |
| Control char rejection | Before build | Env values with null bytes, newlines, or carriage returns are rejected |
| Path traversal check | After extraction (precompiled) | Extracted paths must not contain `..` or absolute paths |
| Redirect safety | During download | HTTP client blocks redirects to non-HTTPS URLs |
| Size limit | During download | Downloads capped at 2 GB |
