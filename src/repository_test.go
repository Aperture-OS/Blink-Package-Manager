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

func TestFindRepoForPackage_Found(t *testing.T) {
	tmp := t.TempDir()

	origLocalRepo := LocalRepositoryDirPath
	defer func() { LocalRepositoryDirPath = origLocalRepo }()
	LocalRepositoryDirPath = tmp

	repoDir := filepath.Join(tmp, "testrepo", "recipes")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "mypkg.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	repos := map[string]RepoConfig{
		"testrepo": {
			Name: "testrepo",
			URL:  "https://github.com/example/repo.git",
			Ref:  "main",
		},
	}

	repo, path, err := FindRepoForPackage("mypkg", repos)
	if err != nil {
		t.Fatalf("FindRepoForPackage failed: %v", err)
	}
	if repo.Name != "testrepo" {
		t.Errorf("expected repo name 'testrepo', got %q", repo.Name)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestFindRepoForPackage_NotFound(t *testing.T) {
	tmp := t.TempDir()

	origLocalRepo := LocalRepositoryDirPath
	defer func() { LocalRepositoryDirPath = origLocalRepo }()
	LocalRepositoryDirPath = tmp

	repos := map[string]RepoConfig{
		"testrepo": {
			Name: "testrepo",
			URL:  "https://github.com/example/repo.git",
			Ref:  "main",
		},
	}

	_, _, err := FindRepoForPackage("nonexistent", repos)
	if err == nil {
		t.Fatal("expected error for package not found")
	}
}

func TestFindRepoForPackage_EmptyRepos(t *testing.T) {
	repos := map[string]RepoConfig{}

	_, _, err := FindRepoForPackage("anypkg", repos)
	if err == nil {
		t.Fatal("expected error for empty repos")
	}
}
