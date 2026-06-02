# CLI Usage

Blink provides a simple command-line interface for managing packages. All commands that modify the system require root privileges.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--root` | `-r` | Specify an alternative root directory (default: `/`) |
| `--force` | `-f` | Force the operation (re-download, reinstall, etc.) |
| `--path` | `-p` | Specify an alternative recipes directory |

## Commands

### Install a Package

```bash
sudo blink install <package>
```

Downloads the recipe, resolves dependencies, downloads source, verifies SHA256, builds, and installs the package.

**Example:**
```
$ sudo blink install hello
Blink Package Manager Version: v0.2.0-alpha
© Copyright 2025-2026 Aperture OS. All rights reserved.
[BLINK] 14:30:01 INFO: Processing package: hello
[BLINK] 14:30:01 INFO: Fetching package "hello"
Repository: "main" (https://github.com/Aperture-OS/testing-blink-repo.git)
Name       :        hello
Version    :     1.0.0
Release    :     1768153997
Description: Hello World program
Author     :      example.com
License    :     MIT
[BLINK] 14:30:05 INFO: Decompressing source for hello into /var/blink/build/hello
[BLINK] 14:30:06 SUCCESS: Package hello installed successfully
```

### Uninstall a Package

```bash
sudo blink uninstall <package>
```

Runs the package's uninstall commands and removes it from the manifest.

**Aliases:** `remove`, `u`, `uninst`

### Search / Package Info

```bash
sudo blink search <package>
```

Fetches and displays detailed information about a package without installing it.

**Aliases:** `info`, `f`, `fetch`, `details`, `pkginfo`, `information`, `searchfor`

**Example:**
```
$ sudo blink search vim
Repository: "main" (https://github.com/Aperture-OS/testing-blink-repo.git)
Name       :        vim
Version    :     9.1.0
Release    :     1768200000
Description: Vi IMproved - enhanced vi editor
Author     :      vim.org
License    :     Vim
```

### Download a Recipe

```bash
sudo blink get <package>
```

Downloads only the package recipe (JSON file) without installing.

**Aliases:** `d`, `download`, `g`, `dl`

### Sync Repositories

```bash
sudo blink sync
```

Clones or updates all configured package repositories to the latest version.

**Aliases:** `s`, `--sync`, `repo`, `reposync`

### Update All Packages

```bash
sudo blink update
```

Checks all installed packages for updates and prompts to upgrade those with a newer release in the repository.

**Aliases:** `upgrade`, `up`

**Example:**
```
$ sudo blink update
[BLINK] 14:35:00 INFO: Syncing repositories...
[BLINK] 14:35:02 INFO: Update available: hello (1768153997 → 1768200000)
[BLINK] 14:35:02 INFO: Up to date: vim
[BLINK] 14:35:02 WARN: Packages to update: 1
 - hello
[BLINK] 14:35:02 WARN: Proceed with update? [ (Y)es / (N)o ]:
```

### Clean Cache

```bash
sudo blink clean
```

Removes cached recipes, downloaded sources, and build artifacts.

**Aliases:** `cleanup`, `clear`, `c`, `-c`, `--clean`, `--cleanup`

### Show Version

```bash
blink version
```

Displays the current Blink version, license, and copyright information.

**Aliases:** `v`, `ver`, `--version`, `-v`

### Show Support Info

```bash
blink support
```

Displays links to Discord and GitHub Issues for getting help.

**Aliases:** `issue`, `bug`, `contact`, `discord`, `--support`, `--bug`

### Generate Shell Completions

```bash
blink completion bash
blink completion zsh
blink completion fish
```

Generates shell completion scripts. Pipe to a file to install:

```bash
blink completion bash > /etc/bash_completion.d/blink
blink completion zsh > /usr/local/share/zsh/site-functions/_blink
blink completion fish > ~/.config/fish/completions/blink.fish
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid arguments, failed operations) |
| Non-zero | Fatal error (permission denied, lock contention, etc.) |
