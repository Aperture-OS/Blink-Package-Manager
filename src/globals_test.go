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
	"path/filepath"
	"testing"
)

func TestComputePaths(t *testing.T) {
	paths := ComputePaths("/myroot")

	expected := filepath.Join("/myroot", "var", "blink")
	if paths.BaseDataDir != expected {
		t.Errorf("BaseDataDir = %q, want %q", paths.BaseDataDir, expected)
	}

	expectedConfig := filepath.Join(expected, "etc", "config.toml")
	if paths.ConfigFile != expectedConfig {
		t.Errorf("ConfigFile = %q, want %q", paths.ConfigFile, expectedConfig)
	}

	expectedLock := filepath.Join(expected, "etc", "blink.lock")
	if paths.LockFile != expectedLock {
		t.Errorf("LockFile = %q, want %q", paths.LockFile, expectedLock)
	}

	expectedRepo := filepath.Join(expected, "repositories")
	if paths.LocalRepoDir != expectedRepo {
		t.Errorf("LocalRepoDir = %q, want %q", paths.LocalRepoDir, expectedRepo)
	}

	expectedSource := filepath.Join(expected, "sources")
	if paths.SourceDir != expectedSource {
		t.Errorf("SourceDir = %q, want %q", paths.SourceDir, expectedSource)
	}

	expectedRecipe := filepath.Join(expected, "recipes")
	if paths.RecipeDir != expectedRecipe {
		t.Errorf("RecipeDir = %q, want %q", paths.RecipeDir, expectedRecipe)
	}

	expectedManifest := filepath.Join(expected, "etc", "manifest.toml")
	if paths.ManifestFile != expectedManifest {
		t.Errorf("ManifestFile = %q, want %q", paths.ManifestFile, expectedManifest)
	}

	expectedBuild := filepath.Join(expected, "build")
	if paths.BuildDir != expectedBuild {
		t.Errorf("BuildDir = %q, want %q", paths.BuildDir, expectedBuild)
	}
}

func TestApplyRoot_Valid(t *testing.T) {
	tmp := t.TempDir()

	origBase := BaseDataDirPath
	origConfig := ConfigFilePath
	origLock := LockFilePath
	origLocalRepo := LocalRepositoryDirPath
	origSource := SourceDirPath
	origRecipe := RecipeDirPath
	origManifest := ManifestFilePath
	origBuild := BuildDirPath
	defer func() {
		BaseDataDirPath = origBase
		ConfigFilePath = origConfig
		LockFilePath = origLock
		LocalRepositoryDirPath = origLocalRepo
		SourceDirPath = origSource
		RecipeDirPath = origRecipe
		ManifestFilePath = origManifest
		BuildDirPath = origBuild
	}()

	err := ApplyRoot(tmp)
	if err != nil {
		t.Fatalf("ApplyRoot failed: %v", err)
	}

	expectedBase := filepath.Join(tmp, "var", "blink")
	if BaseDataDirPath != expectedBase {
		t.Errorf("BaseDataDirPath = %q, want %q", BaseDataDirPath, expectedBase)
	}
}

func TestApplyRoot_EmptyPath(t *testing.T) {
	err := ApplyRoot("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestApplyRoot_RelativePath(t *testing.T) {
	err := ApplyRoot("relative/path")
	if err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestValidatePackageName_Valid(t *testing.T) {
	names := []string{"vim", "gcc", "linux-headers", "lib2to3", "my.pkg", "pkg+extra"}
	for _, n := range names {
		if err := validatePackageName(n); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", n, err)
		}
	}
}

func TestValidatePackageName_Invalid(t *testing.T) {
	names := []string{"", ".hidden", "-dash", "with spaces", "a/b", "pkg;evil", "$(cmd)"}
	for _, n := range names {
		if err := validatePackageName(n); err == nil {
			t.Errorf("expected %q to be invalid, got nil", n)
		}
	}
}

func TestValidateRoot_Valid(t *testing.T) {
	valid := []string{"/opt", "/home/user", "/mnt/data"}
	for _, r := range valid {
		if err := validateRoot(r); err != nil {
			t.Errorf("expected %q to be valid root, got error: %v", r, err)
		}
	}
}

func TestValidateRoot_ForbiddenPaths(t *testing.T) {
	forbidden := []string{"/", "/proc", "/sys", "/dev", "/run"}
	for _, r := range forbidden {
		if err := validateRoot(r); err == nil {
			t.Errorf("expected %q to be rejected as root, got nil", r)
		}
	}
}

func TestValidateRoot_RelativePath(t *testing.T) {
	if err := validateRoot("relative"); err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestGetPaths_ReflectsGlobals(t *testing.T) {
	origBase := BaseDataDirPath
	defer func() { BaseDataDirPath = origBase }()

	BaseDataDirPath = "/test/blink"
	p := getPaths()
	if p.BaseDataDir != "/test/blink" {
		t.Errorf("getPaths().BaseDataDir = %q, want %q", p.BaseDataDir, "/test/blink")
	}
}
