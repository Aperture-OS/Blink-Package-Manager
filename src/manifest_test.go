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
	"os"
	"path/filepath"
	"testing"
)

func setupTestPaths(t *testing.T) (string, func()) {
	t.Helper()
	tmp := t.TempDir()

	origBase := BaseDataDirPath
	origConfig := ConfigFilePath
	origLock := LockFilePath
	origLocalRepo := LocalRepositoryDirPath
	origSource := SourceDirPath
	origRecipe := RecipeDirPath
	origManifest := ManifestFilePath
	origBuild := BuildDirPath

	paths := ComputePaths(tmp)
	BaseDataDirPath = paths.BaseDataDir
	ConfigFilePath = paths.ConfigFile
	LockFilePath = paths.LockFile
	LocalRepositoryDirPath = paths.LocalRepoDir
	SourceDirPath = paths.SourceDir
	RecipeDirPath = paths.RecipeDir
	ManifestFilePath = paths.ManifestFile
	BuildDirPath = paths.BuildDir

	for _, d := range []string{
		filepath.Dir(paths.ConfigFile),
		paths.LocalRepoDir,
		paths.RecipeDir,
		paths.SourceDir,
		paths.BuildDir,
	} {
		os.MkdirAll(d, 0750)
	}

	cleanup := func() {
		BaseDataDirPath = origBase
		ConfigFilePath = origConfig
		LockFilePath = origLock
		LocalRepositoryDirPath = origLocalRepo
		SourceDirPath = origSource
		RecipeDirPath = origRecipe
		ManifestFilePath = origManifest
		BuildDirPath = origBuild
	}

	return tmp, cleanup
}

func TestEnsureManifest_CreatesFile(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatalf("ensureManifest failed: %v", err)
	}

	if _, err := os.Stat(ManifestFilePath); os.IsNotExist(err) {
		t.Fatal("manifest file was not created")
	}
}

func TestLoadManifest_Empty(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	m, err := loadManifest()
	if err != nil {
		t.Fatalf("loadManifest failed: %v", err)
	}

	if len(m.Installed) != 0 {
		t.Fatalf("expected empty manifest, got %d installed packages", len(m.Installed))
	}
}

func TestAddToManifest_And_ManifestHas(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageInfo{
		Name:    "testpkg",
		Version: "1.0.0",
		Release: 1000,
	}

	if err := addToManifest(pkg); err != nil {
		t.Fatalf("addToManifest failed: %v", err)
	}

	installed, exists, err := manifestHas("testpkg")
	if err != nil {
		t.Fatalf("manifestHas error: %v", err)
	}
	if !exists {
		t.Fatal("expected package to exist in manifest")
	}
	if installed.Name != "testpkg" {
		t.Errorf("expected name 'testpkg', got %q", installed.Name)
	}
	if installed.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", installed.Version)
	}
}

func TestAddToManifest_DuplicateNoError(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageInfo{Name: "dup", Version: "1.0.0", Release: 1000}

	if err := addToManifest(pkg); err != nil {
		t.Fatal(err)
	}

	if err := addToManifest(pkg); err != nil {
		t.Fatalf("adding duplicate should not error: %v", err)
	}

	m, err := loadManifest()
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, p := range m.Installed {
		if p.Name == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 entry for 'dup', got %d", count)
	}
}

func TestRemoveFromManifest(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageInfo{Name: "removeme", Version: "2.0.0", Release: 2000}

	if err := addToManifest(pkg); err != nil {
		t.Fatal(err)
	}

	if err := removeFromManifest(pkg); err != nil {
		t.Fatalf("removeFromManifest failed: %v", err)
	}

	_, exists, err := manifestHas("removeme")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected package to be removed from manifest")
	}
}

func TestRemoveFromManifest_NotPresent(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatal(err)
	}

	pkg := PackageInfo{Name: "ghost", Version: "1.0.0"}

	err := removeFromManifest(pkg)
	if err != nil {
		t.Fatalf("removing non-existent package should not error: %v", err)
	}
}

func TestIsInstalled(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	if err := ensureManifest(); err != nil {
		t.Fatal(err)
	}

	if isInstalled("nonexistent") {
		t.Fatal("expected nonexistent package to not be installed")
	}

	pkg := PackageInfo{Name: "installed-pkg", Version: "1.0.0", Release: 100}
	if err := addToManifest(pkg); err != nil {
		t.Fatal(err)
	}

	if !isInstalled("installed-pkg") {
		t.Fatal("expected installed-pkg to be installed")
	}
}

func TestSaveManifest_AtomicWrite(t *testing.T) {
	_, cleanup := setupTestPaths(t)
	defer cleanup()

	m := Manifest{
		Installed: []InstalledPkg{
			{Name: "pkg1", Version: "1.0.0", Release: 100},
			{Name: "pkg2", Version: "2.0.0", Release: 200},
		},
	}

	if err := saveManifest(m); err != nil {
		t.Fatalf("saveManifest failed: %v", err)
	}

	loaded, err := loadManifest()
	if err != nil {
		t.Fatalf("loadManifest after save failed: %v", err)
	}

	if len(loaded.Installed) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(loaded.Installed))
	}
}
