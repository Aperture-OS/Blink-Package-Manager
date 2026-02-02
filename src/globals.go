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
	"time"
)

// Paths holds all the computed paths for a given root directory
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

// ComputePaths computes all paths based on a root directory
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

// ApplyRoot validates the provided root path, computes derived paths
// and applies them to the global variables used across the program.
// This avoids mutating globals in ad-hoc ways and centralizes path setup.
func ApplyRoot(osRoot string) error {
	if osRoot == "" || osRoot == "\n" {
		osRoot = "/"
	}

	cleaned := filepath.Clean(osRoot)
	if !filepath.IsAbs(cleaned) {
		abs, err := filepath.Abs(cleaned)
		if err != nil {
			return fmt.Errorf("failed to resolve root path: %v", err)
		}
		cleaned = abs
	}

	paths := ComputePaths(cleaned) // <- just use cleaned root

	// Create base + subdirs
	subdirs := []string{
		filepath.Dir(paths.ConfigFile), // etc
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

	// Apply globals
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
//							     Globals
//===================================================================//

// !!! NEVER USE RELATIVE PATHS FOR GLOBALS !!!
// ALWAYS USE ABSOLUTE PATHS TO AVOID ISSUES
// EVEN WHEN TESTING LOCALLY, ALWAYS ABSOLUTE PATHS!

var (
	DistroName = "ApertureOS"

	BaseDataDirPath     = "/var/blink" // Default: /var/blink
	CurrentYear         = time.Now().Year()                               // Current year for copyright
	CurrentBlinkVersion = "v0.2.0-alpha"                                  // Blink version

	DefaultRepositoryList = `
[pseudoRepository]
git_url = "https://github.com/Aperture-OS/testing-blink-repo.git"
branch = "main"
`

	DefaultRoot = "/" // Default root directory

	ConfigFilePath         = filepath.Join(BaseDataDirPath, "etc", "config.toml")
	LockFilePath           = filepath.Join(BaseDataDirPath, "etc", "blink.lock") // Path to lock file
	LocalRepositoryDirPath = filepath.Join(BaseDataDirPath, "repositories")
	SourceDirPath          = filepath.Join(BaseDataDirPath, "sources") // Path to downloaded source
	RecipeDirPath          = filepath.Join(BaseDataDirPath, "recipes")
	ManifestFilePath       = filepath.Join(BaseDataDirPath, "etc", "manifest.toml")
	BuildDirPath           = filepath.Join(BaseDataDirPath, "build")

	lock = &Lock{Path: LockFilePath}

	SupportInformationSnippet = // Support information string
	`Having trouble? Join our Discord Server or open a GitHub issue.
	Include any DEBUG INFO logs when reporting issues.
	Discord: https://discord.com/invite/rx82u93hGD
	GitHub Issues: https://github.com/Aperture-OS/blink-package-manager/issues`

	VersionInformationSnippet = // version information string
	fmt.Sprintf(`Blink Package Manager - Version %s 
	Licensed under GPL v3.0 by Aperture OS
	https://aperture-os.github.io
	All rights reserved. © Copyright 2025-%d Aperture OS.
	`, CurrentBlinkVersion, CurrentYear)
) // TODO: migrate to /var/blink
