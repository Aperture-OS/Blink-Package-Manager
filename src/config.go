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
  along with this program.  If not, see <https://www.apache.org/licenses/LICENSE-2.0>.
*/

package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Aperture-OS/eyes"
	"github.com/BurntSushi/toml"
)

const defaultConfigTOML = `[pseudoRepository]
git_url = "https://github.com/Aperture-OS/testing-blink-repo.git"
branch = "main"
trustedKey = "/etc/blink/trusted.pub"
`

var validBranchName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9/_.-]{0,199}$`)

func validateRepoEntry(name, rawURL, branch string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("repo %q has unparseable git_url: %v", name, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("repo %q: only https:// git URLs are permitted, got scheme %q", name, u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("repo %q: git_url must have a non-empty host", name)
	}
	if !validBranchName.MatchString(branch) {
		return fmt.Errorf("repo %q: branch name %q contains disallowed characters", name, branch)
	}
	return nil
}

func writeConfigAtomic(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(defaultConfigTOML); err != nil {
		return fmt.Errorf("failed to write default config: %v", err)
	}
	return nil
}

func CreateDefaultConfig() error {
	if ConfigFilePath == "" {
		return fmt.Errorf("ConfigFilePath is empty")
	}
	if err := os.MkdirAll(filepath.Dir(ConfigFilePath), 0750); err != nil {
		return err
	}
	if err := writeConfigAtomic(ConfigFilePath); err != nil {
		return err
	}
	eyes.Infof("Default repository config created at %s", ConfigFilePath)
	return nil
}

func EnsureConfig() error {
	if err := os.MkdirAll(filepath.Dir(ConfigFilePath), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %v", err)
	}
	if err := writeConfigAtomic(ConfigFilePath); err != nil {
		return err
	}
	return nil
}

func LoadConfig() (map[string]RepoConfig, error) {
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		eyes.Infof("Config file not found. Creating default config at %s", ConfigFilePath)
		if err := CreateDefaultConfig(); err != nil {
			return nil, err
		}
	}

	var repos map[string]RepoConfig
	if _, err := toml.DecodeFile(ConfigFilePath, &repos); err != nil {
		return nil, fmt.Errorf("failed to decode config TOML: %v", err)
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in config")
	}

	for name, repo := range repos {
		if err := validateRepoEntry(name, repo.URL, repo.Ref); err != nil {
			return nil, err
		}
	}

	eyes.Infof("Loaded %d repositories from %s", len(repos), ConfigFilePath)
	return repos, nil
}

func LoadRepos(path string) (map[string]RepoConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("repo config does not exist: %s", path)
	}

	var raw map[string]struct {
		GitURL string `toml:"git_url"`
		Branch string `toml:"branch"`
		Hash   string `toml:"hash"`
		Key    string `toml:"trusted_key"`
	}

	if _, err := toml.DecodeFile(path, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode repo config: %v", err)
	}

	repos := make(map[string]RepoConfig)
	for name, r := range raw {
		if err := validateRepoEntry(name, r.GitURL, r.Branch); err != nil {
			return nil, err
		}
		repos[name] = RepoConfig{
			Name:       name,
			URL:        r.GitURL,
			Ref:        r.Branch,
			Hash:       r.Hash,
			TrustedKey: r.Key,
		}
	}

	return repos, nil
}
