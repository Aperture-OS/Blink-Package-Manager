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

func TestValidateRepoEntry_Valid(t *testing.T) {
	err := validateRepoEntry("testrepo", "https://github.com/example/repo.git", "main")
	if err != nil {
		t.Fatalf("expected valid repo entry, got error: %v", err)
	}
}

func TestValidateRepoEntry_HTTPRejected(t *testing.T) {
	err := validateRepoEntry("testrepo", "http://github.com/example/repo.git", "main")
	if err == nil {
		t.Fatal("expected error for http:// URL, got nil")
	}
}

func TestValidateRepoEntry_EmptyHost(t *testing.T) {
	err := validateRepoEntry("testrepo", "https:///repo.git", "main")
	if err == nil {
		t.Fatal("expected error for empty host, got nil")
	}
}

func TestValidateRepoEntry_InvalidBranch(t *testing.T) {
	err := validateRepoEntry("testrepo", "https://github.com/example/repo.git", "bad branch!")
	if err == nil {
		t.Fatal("expected error for invalid branch name, got nil")
	}
}

func TestValidateRepoEntry_ValidBranches(t *testing.T) {
	branches := []string{"main", "develop", "release/v1.0", "feature/my-feature", "v1.0.0"}
	for _, b := range branches {
		err := validateRepoEntry("testrepo", "https://github.com/example/repo.git", b)
		if err != nil {
			t.Errorf("expected branch %q to be valid, got error: %v", b, err)
		}
	}
}

func TestWriteConfigAtomic_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.toml")

	err := writeConfigAtomic(cfgPath)
	if err != nil {
		t.Fatalf("writeConfigAtomic failed: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}

	if string(data) != defaultConfigTOML {
		t.Fatalf("config content mismatch.\ngot:  %q\nwant: %q", string(data), defaultConfigTOML)
	}
}

func TestWriteConfigAtomic_DoesNotOverwrite(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.toml")

	existing := "existing content"
	if err := os.WriteFile(cfgPath, []byte(existing), 0640); err != nil {
		t.Fatal(err)
	}

	err := writeConfigAtomic(cfgPath)
	if err != nil {
		t.Fatalf("writeConfigAtomic should not error on existing file: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != existing {
		t.Fatal("writeConfigAtomic overwrote existing file")
	}
}

func TestLoadRepos_ValidConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "repos.toml")

	content := `[myrepo]
git_url = "https://github.com/example/repo.git"
branch = "main"
hash = "abc123"
trusted_key = ""
`
	if err := os.WriteFile(cfgPath, []byte(content), 0640); err != nil {
		t.Fatal(err)
	}

	repos, err := LoadRepos(cfgPath)
	if err != nil {
		t.Fatalf("LoadRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	repo, ok := repos["myrepo"]
	if !ok {
		t.Fatal("expected repo 'myrepo' to exist")
	}
	if repo.URL != "https://github.com/example/repo.git" {
		t.Errorf("unexpected URL: %s", repo.URL)
	}
	if repo.Ref != "main" {
		t.Errorf("unexpected Ref: %s", repo.Ref)
	}
	if repo.Hash != "abc123" {
		t.Errorf("unexpected Hash: %s", repo.Hash)
	}
}

func TestLoadRepos_MissingFile(t *testing.T) {
	_, err := LoadRepos("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadRepos_InvalidTOML(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.toml")

	if err := os.WriteFile(cfgPath, []byte("this is not valid toml [[["), 0640); err != nil {
		t.Fatal(err)
	}

	_, err := LoadRepos(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestLoadRepos_InvalidURL(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "repos.toml")

	content := `[badrepo]
git_url = "http://insecure.com/repo.git"
branch = "main"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0640); err != nil {
		t.Fatal(err)
	}

	_, err := LoadRepos(cfgPath)
	if err == nil {
		t.Fatal("expected error for http:// URL")
	}
}

func TestLoadConfig_CreatesDefaultWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	origPath := ConfigFilePath
	defer func() { ConfigFilePath = origPath }()

	ConfigFilePath = filepath.Join(tmp, "etc", "config.toml")

	repos, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(repos) == 0 {
		t.Fatal("expected at least one repo from default config")
	}

	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created")
	}
}
