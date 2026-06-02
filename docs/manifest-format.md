# Manifest Format Specification

Blink uses a TOML-based manifest file to track installed packages. The manifest is located at:

```
/var/blink/etc/manifest.toml
```

When using a custom root (`--root /myroot`), the path becomes:

```
/myroot/var/blink/etc/manifest.toml
```

## Structure

The manifest contains a single top-level array of tables called `installed`:

```toml
[[installed]]
name = "hello"
version = "1.0.0"
release = 1768153997

[[installed]]
name = "vim"
version = "9.1.0"
release = 1768200000
```

## Fields

### `name`
- **Type:** String
- **Required:** Yes
- The unique package identifier matching the recipe filename (without `.json`).

### `version`
- **Type:** String
- **Required:** Yes
- The upstream version of the installed package.

### `release`
- **Type:** Integer (int64)
- **Required:** Yes
- Unix timestamp of the package recipe's last update. Used to determine if updates are available by comparing with the repository's recipe release value.

## Behavior

- The manifest is created automatically on first install if it does not exist.
- Writes are atomic: Blink writes to a `.tmp` file, syncs it to disk, then renames it over the original.
- Duplicate entries are prevented; adding a package that already exists is a no-op.
- Removing a package that is not in the manifest is a no-op (no error).

## Configuration File

Blink's repository configuration is a TOML file at:

```
/var/blink/etc/config.toml
```

### Structure

Each section defines a repository:

```toml
[myrepo]
git_url = "https://github.com/Aperture-OS/testing-blink-repo.git"
branch = "main"
hash = ""
trustedKey = "/etc/blink/trusted.pub"
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `git_url` | String | HTTPS URL of the Git repository. Only `https://` is permitted. |
| `branch` | String | Git branch to track. Must match `^[a-zA-Z0-9][a-zA-Z0-9/_.-]{0,199}$`. |
| `hash` | String | Optional pinned commit hash. If set, Blink verifies the resolved commit starts with this prefix. |
| `trustedKey` | String | Path to the GPG public key used to verify commit signatures. |
