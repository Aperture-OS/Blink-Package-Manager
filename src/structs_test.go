/*
  Blink, a powerful source-based package manager. Core of ApertureOS.
  Want to use it for your own project?
  Blink is completely FOSS (Free and Open Source),
  edit, publish, use, contribute to Blink however you prefer.
  Copyright (C) 2025-2026 Aperture OS

  This program is free software: you can redistribute it and/or modify
  it under the terms of the Apache 2.0 License as published by
  the Apache Software Foundation, either version 2.0 of the License, or
  any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

  You should have received a copy of the Apache 2.0 License
  along with this program. If not, see <https://www.apache.org/licenses/LICENSE-2.0>.
*/

package main

import (
	"testing"
)

func TestPackageInfoValidate_ValidPackage(t *testing.T) {
	pkg := PackageInfo{
		Name:    "testpkg",
		Version: "1.0.0",
	}
	pkg.Source.URL = "https://example.com/testpkg-1.0.0.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"

	if err := pkg.Validate(); err != nil {
		t.Fatalf("expected valid package, got error: %v", err)
	}
}

func TestPackageInfoValidate_HTTPSchemeRejected(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "http://example.com/testpkg-1.0.0.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for http:// URL, got nil")
	}
}

func TestPackageInfoValidate_EmptyHost(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https:///path/to/file.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for empty host, got nil")
	}
}

func TestPackageInfoValidate_InvalidSha256(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https://example.com/testpkg.tar.gz"
	pkg.Source.Sha256 = "tooshort"
	pkg.Source.Type = "tar.gz"

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid sha256, got nil")
	}
}

func TestPackageInfoValidate_UnsupportedSourceType(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https://example.com/testpkg.rar"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "rar"

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for unsupported source type, got nil")
	}
}

func TestPackageInfoValidate_AllowedSourceTypes(t *testing.T) {
	types := []string{"tar.gz", "tar.xz", "tar.bz2", "tar.zst", "zip"}
	for _, st := range types {
		pkg := PackageInfo{}
		pkg.Source.URL = "https://example.com/testpkg.archive"
		pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
		pkg.Source.Type = st

		if err := pkg.Validate(); err != nil {
			t.Errorf("expected source type %q to be valid, got error: %v", st, err)
		}
	}
}

func TestPackageInfoValidate_DisallowedBuildEnvKey(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https://example.com/testpkg.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"
	pkg.Build.Env = map[string]string{"MALICIOUS_VAR": "value"}

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for disallowed env key, got nil")
	}
}

func TestPackageInfoValidate_AllowedBuildEnvKeys(t *testing.T) {
	keys := []string{"CC", "CXX", "CFLAGS", "CXXFLAGS", "LDFLAGS", "PREFIX", "DESTDIR", "MAKEFLAGS", "PKG_CONFIG_PATH"}
	for _, k := range keys {
		pkg := PackageInfo{}
		pkg.Source.URL = "https://example.com/testpkg.tar.gz"
		pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
		pkg.Source.Type = "tar.gz"
		pkg.Build.Env = map[string]string{k: "somevalue"}

		if err := pkg.Validate(); err != nil {
			t.Errorf("expected env key %q to be allowed, got error: %v", k, err)
		}
	}
}

func TestPackageInfoValidate_EnvValueWithControlChars(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https://example.com/testpkg.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"
	pkg.Build.Env = map[string]string{"CC": "gcc\x00injected"}

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for env value with null byte, got nil")
	}
}

func TestPackageInfoValidate_EnvValueWithNewline(t *testing.T) {
	pkg := PackageInfo{}
	pkg.Source.URL = "https://example.com/testpkg.tar.gz"
	pkg.Source.Sha256 = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	pkg.Source.Type = "tar.gz"
	pkg.Build.Env = map[string]string{"CC": "gcc\ninjected"}

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected error for env value with newline, got nil")
	}
}
