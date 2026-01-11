<p align="center">
<br/>
<img src="https://github.com/Aperture-OS/branding/blob/main/Logo-Bright/logo-bright.png?raw=true" alt="ApertureOS Logo" width=100>
<h3 align="center">Contributing to Blink</h3>
<p align="center">
Here you will find the instructions on how to Contribute to Blink, its Package Repository, and more.</p>
<p align="center">If you are working on a repository, it is recommended you leave this file in the root of the project.
</p>
<p align="center">
<img alt="Static Badge" src="https://img.shields.io/badge/Blink_Package_Manager-v0.1.0-6d7592?logo=github&labelColor=45455e&link=https%3A%2F%2Fgithub.com%2FAperture-OS%2Fblink-package-manager%2Freleases">
</p>

# The Package Repository

The package repository is a repository on GitHub, GitLab, CodeBerg, or a http(s) mirror (http(s) Mirrors aren't recommended), containing all the package recipes (.json files with all the data for a package to be installed), it's root ( / ) directory looks something like:
```tree
.
├── repositoryName1
│   ├── package.json
│   ├── package2.json
│   ├── package3.json
│   └── etc.json
├── repositoryName2
│   ├── package4.json
│   ├── package5.json
│   ├── package6.json
│   └── etcetera.json  
├── README.md (optional)
├── LICENSE (optional)
└── CONTRIBUTING.md (recommended, this file)
```

`package.json` would look something like this:

```json
{
  "name": "package",
  "version": "1.0.0",
  "release": 1768153997,
  "description": "Package is a package.",
  "author": "example.com",
  "license": "MIT",
  "source": {
    "url": "https://example.com/package.tar.gz",
    "sha256": "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
    "type": "tar.gz"
  },
  "dependencies": {
    "dependency1": ">=1.0.0",
	"dependency2": ">=1.0.0"
  },
  "opt_dependencies": [
    {
      "id": 1,
      "description": "This gets shown in the optional dependency resolution for group id: 1.",
      "options": [
        "opt-dep1",
		"opt-dep2",
		"opt-dep3",
      ]
    }
  ],

  "build": {
    "kind": "toCompile", // or preCompiled
    "env": {
      "MAKEFLAGS": "-j$(nproc)"
    },
    "prepare": ["rm -rf ~/.cache/test"],
    "install": ["make install PREFIX=${PREFIX:-/usr/local}"],
    "uninstall": ["make uninstall PREFIX=${PREFIX:-/usr/local}"]
  }
}
```

## 1. Package Metadata

```json
{
  "name": "package",
  "version": "1.0.0",
  "release": 1768153997,
  "description": "Package is a package.",
  "author": "example.com",
  "license": "MIT",
````

### `name`

* The **unique identifier** of the package.
* This is what users will install (`blink install package`).
* **\*\*\* MUST BE THE NAME OF THE FILE WITHOUT .json \*\*\***
* Should be lowercase, consistent, and stable.

### `version`

* The upstream software version.
* Used for dependency resolution and upgrades.
* Follows semantic versioning (`major.minor.patch`) when possible.

### `release`

* A Unix timestamp determining when the package was last updated.
* Used to distinguish multiple package builds of the same version.
* Useful alias:

```sh
alias now='date +%s'
alias copynow='date +%s | wl-copy' # for wayland
```

### `description`

* Short, human-readable summary of the package.
* Shown in search results and package info.

### `author`

* The upstream author or project homepage.
* Can be a name, URL, or organization.
* Usually the GitHub/GitLab/CodeBerg page of the author

### `license`

* The software license identifier.
* Should be SPDX-compatible when possible (`MIT`, `GPL-3.0`, etc).


## 2. Source Definition

```json
  "source": {
    "url": "https://example.com/package.tar.gz",
    "sha256": "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
    "type": "tar.gz"
  },
```

If it's a precompiled package use this as the precompiled archive definition.

### `url`

* Location of the source archive or binary.
* Can point to GitHub releases, mirrors, or custom servers.
* **MUST BE RAW FILE DOWNLOAD (Compatible with curl/wget)**

### `sha256`

* Cryptographic hash of the downloaded file.
* Ensures **integrity and security**.
* Blink verifies this before building or installing.
* To find it: Either look for one in the release tab of GitHub/Lab/CodeBerg

Or find it manually:

```sh
sha256sum sourcefile.tar.gz # or any other extension
```

### `type`

* Archive or file type.
* Used to determine how Blink extracts or handles the source.
* Supported values: `.tar.gz`, `.tar.bz2`, `.tar.xs`, `.zip`


## 3. Dependencies

```json
  "dependencies": {
    "dependency1": ">=1.0.0",
    "dependency2": ">=1.0.0"
  },
```

### `dependencies`

* Required packages that **must be installed**.
* Version constraints are supported.
* Blink resolves and installs these automatically.

Examples:

* `>=1.0.0` -> at least version 1.0.0
* `=2.1.3` -> exact version
* `<3.0.0` -> any version below 3.0.0


## 4. Optional Dependencies

```json
  "opt_dependencies": [
    {
      "id": 1,
      "description": "This gets shown in the optional dependency resolution for group id: 1.",
      "options": [
        "opt-dep1",
        "opt-dep2",
        "opt-dep3"
      ]
    }
  ],
```

### `opt_dependencies`

* Non-essential features users can choose to install.
* Presented interactively or via flags.

#### `id`

* Group identifier.
* Allows Blink to group related optional dependencies together.

#### `description`

* Shown to the user when resolving optional dependencies.
* Explains *why* these options exist.

#### `options`

* List of package names the user may select from.
* Installing one or more is optional.


## 5. Build Instructions

```json
  "build": {
```

This section defines **how the package is built, installed, and removed**.


### 5.1 Build Kind

```json
    "kind": "toCompile"
```

* Determines how Blink handles the package.
* Possible values:

  * `toCompile` -> build from source
  * `preCompiled` -> install binaries directly


### 5.2 Build Environment

```json
    "env": {
      "MAKEFLAGS": "-j$(nproc)"
    },
```

* Environment variables used during build.
* Injected into the build process.
* Useful for parallel builds, paths, or compiler flags.


### 5.3 Prepare Step

```json
    "prepare": ["rm -rf ~/.cache/test"],
```

* Commands run **before building**.

* Used to:

  * Clean previous builds
  * Patch files
  * Prepare directories

* Executed in order, line by line.


### 5.4 Install Step

```json
    "install": ["make install PREFIX=${PREFIX:-/usr/local}"],
```

* Commands used to install files into the system or staging directory.
* `${PREFIX}` allows relocatable installs.
* Defaults to `/usr/local` if not provided.


### 5.5 Uninstall Step

```json
    "uninstall": ["make uninstall PREFIX=${PREFIX:-/usr/local}"]
```

* Commands used to remove the package.
* Ensures clean removal without leftovers.
* Optional but strongly recommended.


## 6. Full Lifecycle Summary

1. **Download** source from `url`
2. **Verify** integrity using `sha256`
3. **Resolve dependencies**
4. **Prepare** build environment
5. **Build or extract** depending on `kind`
6. **Install** files to the system
7. **Optionally remove** via uninstall instructions


## Notes & Best Practices

* Keep recipes deterministic and reproducible
* Always include checksums
* Prefer clear version constraints
* Use optional dependencies for feature toggles
* Test install *and* uninstall paths