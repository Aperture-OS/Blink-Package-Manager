# Repository Indexing and Publishing

This document describes how to create, structure, and publish a Blink package repository.

## Repository Structure

A Blink package repository is a Git repository containing package recipes (JSON files):

```
my-blink-repo/
├── recipes/
│   ├── hello.json
│   ├── vim.json
│   ├── gcc.json
│   └── ...
├── key.pub          (GPG public key for commit verification)
├── README.md        (optional)
├── CONTRIBUTING.md  (recommended)
└── LICENSE          (optional)
```

## Creating a Recipe

Each package recipe is a JSON file in the `recipes/` directory. The filename **must** match the package `name` field (e.g., `hello.json` for a package named `hello`).

See the [CONTRIBUTING.md](https://github.com/Aperture-OS/blink-package-manager/blob/main/CONTRIBUTING.md) for the complete recipe format specification.

### Minimal Recipe Example

```json
{
  "name": "hello",
  "version": "2.12",
  "release": 1768153997,
  "description": "GNU Hello - a friendly greeting program",
  "author": "gnu.org",
  "license": "GPL-3.0",
  "source": {
    "url": "https://ftp.gnu.org/gnu/hello/hello-2.12.tar.gz",
    "sha256": "cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab",
    "type": "tar.gz"
  },
  "dependencies": {},
  "opt_dependencies": [],
  "build": {
    "kind": "toCompile",
    "env": {},
    "prepare": ["./configure --prefix=/usr"],
    "install": ["make", "make install DESTDIR=${DESTDIR}"],
    "uninstall": ["make uninstall DESTDIR=${DESTDIR}"]
  }
}
```

## Publishing a Repository

### 1. Create a Git Repository

```bash
mkdir my-blink-repo && cd my-blink-repo
git init
mkdir recipes
```

### 2. Add Recipes

Create JSON files in `recipes/` following the format above.

### 3. Set Up GPG Signing

Generate a GPG key and configure Git to sign commits:

```bash
gpg --full-generate-key
git config --local user.signingkey <YOUR_KEY_ID>
git config --local commit.gpgsign true
```

Export the public key to the repository:

```bash
gpg --armor --export <YOUR_KEY_ID> > key.pub
```

### 4. Push to a Git Host

```bash
git add .
git commit -m "Initial repository with recipes"
git remote add origin https://github.com/yourorg/your-blink-repo.git
git push -u origin main
```

### 5. Configure Blink to Use Your Repository

Add an entry to `/var/blink/etc/config.toml`:

```toml
[yourrepo]
git_url = "https://github.com/yourorg/your-blink-repo.git"
branch = "main"
trustedKey = "/etc/blink/trusted.pub"
```

Then sync:

```bash
sudo blink sync
```

## Updating Recipes

1. Edit the recipe JSON file.
2. Update the `release` field to the current Unix timestamp: `date +%s`.
3. Update the `version` field if the upstream version changed.
4. Update the `sha256` field if the source archive changed: `sha256sum newfile.tar.gz`.
5. Commit with a signed commit and push.

## Multiple Repositories

Blink supports multiple repositories. If a package exists in more than one repository, Blink will prompt the user to select which one to use.

Each repository section in `config.toml` must have a unique name:

```toml
[official]
git_url = "https://github.com/Aperture-OS/blink-repo.git"
branch = "main"
trustedKey = "/etc/blink/official.pub"

[community]
git_url = "https://github.com/community/blink-community-repo.git"
branch = "main"
trustedKey = "/etc/blink/community.pub"
```

## Commit Hash Pinning

For increased security, you can pin a repository to a specific commit:

```toml
[official]
git_url = "https://github.com/Aperture-OS/blink-repo.git"
branch = "main"
hash = "a1b2c3d4e5f6"
trustedKey = "/etc/blink/official.pub"
```

Blink will refuse to sync if the resolved commit does not match the pinned hash prefix.
