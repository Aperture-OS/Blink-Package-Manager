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

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.apache.org/licenses/LICENSE-2.0>.
*/

package main

import (
	"os"
	"path/filepath"

	"github.com/Aperture-OS/eyes"
	"github.com/BurntSushi/toml"
)

/****************************************************/
// CreateDefaultConfig writes the default repository config to configPath
/****************************************************/
func CreateDefaultConfig() error {
	if configPath == "" {
		return os.ErrInvalid
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the config file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Wrap defaultRepoConfig from globals.go
	if err := toml.NewEncoder(file).Encode(defaultRepoConfig); err != nil {
		return err
	}

	eyes.Infof("Default repository config created at %s", configPath)
	return nil
}

/****************************************************/
// LoadConfig loads the repository config from configPath
/****************************************************/
func LoadConfig() (map[string]RepoConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		eyes.Infof("Config file not found. Creating default config at %s", configPath)
		if err := CreateDefaultConfig(); err != nil {
			return nil, err
		}
	}

	var repos map[string]RepoConfig
	if _, err := toml.DecodeFile(configPath, &repos); err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, os.ErrInvalid
	}

	eyes.Infof("Loaded %d repositories from %s", len(repos), configPath)
	return repos, nil
}
