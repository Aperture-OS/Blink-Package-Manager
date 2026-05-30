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
	"os"
	"path/filepath"

	"github.com/Aperture-OS/eyes"
	"github.com/BurntSushi/toml"
)

// CreateDefaultConfig writes the default repository config to configPath
func CreateDefaultConfig() error {
	if ConfigFilePath == "" {
		return fmt.Errorf("ConfigFilePath is empty")
	}

	dir := filepath.Dir(ConfigFilePath)
	if err := os.MkdirAll(dir, 0750); err != nil { // tighter permissions
		return err
	}

	// Write the raw TOML directly
	defaultConfig := `[pseudoRepository]
git_url = "https://github.com/Aperture-OS/testing-blink-repo.git"
branch = "main"
`
	if err := os.WriteFile(ConfigFilePath, []byte(defaultConfig), 0640); err != nil {
		return fmt.Errorf("failed to write default config: %v", err)
	}

	eyes.Infof("Default repository config created at %s", ConfigFilePath)
	return nil
}

func EnsureConfig() error {
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		eyes.Infof("Config file not found. Creating default at %s", ConfigFilePath)
		if err := os.MkdirAll(filepath.Dir(ConfigFilePath), 0750); err != nil {
			return fmt.Errorf("failed to create config dir: %v", err)
		}

		defaultConfig := `[pseudoRepository]
git_url = "https://github.com/Aperture-OS/testing-blink-repo.git"
branch = "main"
trustedKey = "/key.pub"
`
		if err := os.WriteFile(ConfigFilePath, []byte(defaultConfig), 0640); err != nil {
			return fmt.Errorf("failed to write default config: %v", err)
		}
	}
	return nil
}

// LoadConfig loads the repository config from configPath
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

	eyes.Infof("Loaded %d repositories from %s", len(repos), ConfigFilePath)
	return repos, nil
}

// LoadRepos reads repository definitions from a TOML file
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
