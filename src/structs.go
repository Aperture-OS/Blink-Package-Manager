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
	"regexp"
	"strings"
)

var validSHA256Re = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

var allowedBuildEnvKeys = map[string]bool{
	"CC":              true,
	"CXX":             true,
	"CFLAGS":          true,
	"CXXFLAGS":        true,
	"LDFLAGS":         true,
	"PREFIX":          true,
	"DESTDIR":         true,
	"MAKEFLAGS":       true,
	"PKG_CONFIG_PATH": true,
}

var allowedSourceTypes = map[string]bool{
	"tar.gz":  true,
	"tar.xz":  true,
	"tar.bz2": true,
	"tar.zst": true,
	"zip":     true,
}

type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Release     int    `json:"release"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
	Source      struct {
		URL    string `json:"url"`
		Type   string `json:"type"`
		Sha256 string `json:"sha256"`
	} `json:"source"`
	Dependencies map[string]string `json:"dependencies"`
	OptDeps      []struct {
		ID          int      `json:"id"`
		Description string   `json:"description"`
		Options     []string `json:"options"`
		Default     string   `json:"default"`
	} `json:"opt_dependencies"`
	Build struct {
		Kind      string            `json:"kind"`
		Env       map[string]string `json:"env"`
		Prepare   []string          `json:"prepare"`
		Install   []string          `json:"install"`
		Uninstall []string          `json:"uninstall"`
	} `json:"build"`
}

func (p *PackageInfo) Validate() error {
	u, err := url.Parse(p.Source.URL)
	if err != nil {
		return fmt.Errorf("source URL is unparseable: %v", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("source URL must use https://, got scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("source URL must have a non-empty host")
	}

	if !validSHA256Re.MatchString(p.Source.Sha256) {
		return fmt.Errorf("source sha256 %q is not a valid 64-character hex digest", p.Source.Sha256)
	}

	if !allowedSourceTypes[p.Source.Type] {
		return fmt.Errorf("unsupported source type %q", p.Source.Type)
	}

	for k, v := range p.Build.Env {
		if !allowedBuildEnvKeys[k] {
			return fmt.Errorf("disallowed build env key %q", k)
		}
		if strings.ContainsAny(v, "\x00\n\r") {
			return fmt.Errorf("build env value for key %q contains disallowed control characters", k)
		}
	}

	return nil
}

type Manifest struct {
	Installed []InstalledPkg `json:"installed"`
}

type InstalledPkg struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Release int64  `json:"release"`
}

type RepoConfig struct {
	Name       string `toml:"-"`
	URL        string `toml:"git_url"`
	Ref        string `toml:"branch"`
	Hash       string `toml:"hash"`
	TrustedKey string `toml:"trustedKey"`
}
