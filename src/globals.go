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
	"sync"
	"time"
)

type Paths struct {
	BaseDataDir  string
	ConfigFile   string
	LockFile     string
	LocalRepoDir string
	SourceDir    string
	RecipeDir    string
	ManifestFile string
	BuildDir     string
}

func ComputePaths(osRoot string) Paths {
	baseDataDir := filepath.Join(osRoot, "var", "blink")

	return Paths{
		BaseDataDir:  baseDataDir,
		ConfigFile:   filepath.Join(baseDataDir, "etc", "config.toml"),
		LockFile:     filepath.Join(baseDataDir, "etc", "blink.lock"),
		LocalRepoDir: filepath.Join(baseDataDir, "repositories"),
		SourceDir:    filepath.Join(baseDataDir, "sources"),
		RecipeDir:    filepath.Join(baseDataDir, "recipes"),
		ManifestFile: filepath.Join(baseDataDir, "etc", "manifest.toml"),
		BuildDir:     filepath.Join(baseDataDir, "build"),
	}
}

var globalPathsMu sync.RWMutex

func getPaths() Paths {
	globalPathsMu.RLock()
	defer globalPathsMu.RUnlock()
	return Paths{
		BaseDataDir:  BaseDataDirPath,
		ConfigFile:   ConfigFilePath,
		LockFile:     LockFilePath,
		LocalRepoDir: LocalRepositoryDirPath,
		SourceDir:    SourceDirPath,
		RecipeDir:    RecipeDirPath,
		ManifestFile: ManifestFilePath,
		BuildDir:     BuildDirPath,
	}
}

func ApplyRoot(osRoot string) error {
	if osRoot == "" {
		return fmt.Errorf("root path must not be empty")
	}

	cleaned := filepath.Clean(osRoot)
	if !filepath.IsAbs(cleaned) {
		return fmt.Errorf("root must be an absolute path, got: %q", osRoot)
	}

	paths := ComputePaths(cleaned)

	subdirs := []string{
		filepath.Dir(paths.ConfigFile),
		paths.LocalRepoDir,
		paths.RecipeDir,
		paths.SourceDir,
		paths.BuildDir,
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create subdir %s: %v", dir, err)
		}
	}

	globalPathsMu.Lock()
	defer globalPathsMu.Unlock()

	BaseDataDirPath = paths.BaseDataDir
	ConfigFilePath = paths.ConfigFile
	LockFilePath = paths.LockFile
	LocalRepositoryDirPath = paths.LocalRepoDir
	SourceDirPath = paths.SourceDir
	RecipeDirPath = paths.RecipeDir
	ManifestFilePath = paths.ManifestFile
	BuildDirPath = paths.BuildDir

	lock = &Lock{Path: LockFilePath}

	return nil
}

//===================================================================//
//								 Globals
//===================================================================//

var (
	DistroName = "ApertureOS"

	BaseDataDirPath     = "/var/blink"
	CurrentYear         = time.Now().Year()
	CurrentBlinkVersion = "v0.2.0-alpha"

	DefaultRoot = "/"

	ConfigFilePath         = filepath.Join(BaseDataDirPath, "etc", "config.toml")
	LockFilePath           = filepath.Join(BaseDataDirPath, "etc", "blink.lock")
	LocalRepositoryDirPath = filepath.Join(BaseDataDirPath, "repositories")
	SourceDirPath          = filepath.Join(BaseDataDirPath, "sources")
	RecipeDirPath          = filepath.Join(BaseDataDirPath, "recipes")
	ManifestFilePath       = filepath.Join(BaseDataDirPath, "etc", "manifest.toml")
	BuildDirPath           = filepath.Join(BaseDataDirPath, "build")

	lock = &Lock{Path: LockFilePath}

	SupportInformationSnippet = `Having trouble? Join our Discord Server or open a GitHub issue.
	Include any DEBUG INFO logs when reporting issues.
	Discord: https://discord.com/invite/rx82u93hGD
	GitHub Issues: https://github.com/Aperture-OS/blink-package-manager/issues`

	VersionInformationSnippet = fmt.Sprintf(`Blink Package Manager - Version %s
	Licensed under Apache 2.0 by Aperture OS
	https://aperture-os.github.io
	All rights reserved. © Copyright 2025-%d Aperture OS.
	`, CurrentBlinkVersion, CurrentYear)
)
