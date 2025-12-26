/*
  Blink, a powerful source-based package manager. Core of ApertureOS.
	Want to use it for your own project?
	Blink is completely FOSS (Free and Open Source),
	edit, publish, use, contribute to Blink however you prefer.
  Copyright (C) 2025 Aperture OS

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

/****************************************************/
// Manifest creation and dependency handling functions
// a manifest is a JSON file that keeps track of installed packages, their versions, and other metadata
// this is useful for managing installed packages, checking for updates, and handling dependencies
// the manifest will be stored in /var/blink/etc/manifest.json (see variable delarations at the start of the file)
/****************************************************/

func ensureManifest() error {
	log.Printf("INFO: Ensuring manifest exists at %s", manifestPath)

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		m := Manifest{Installed: []InstalledPkg{}}
		data, _ := json.MarshalIndent(m, "", "  ")
		return os.WriteFile(manifestPath, data, 0644)
	}

	return nil
}

func loadManifest() (Manifest, error) {
	log.Printf("DEBUG: loading manifest")

	var m Manifest

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return Manifest{Installed: []InstalledPkg{}}, nil
	}

	f, err := os.Open(manifestPath)
	if err != nil {
		return m, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return m, err
	}

	return m, nil
}

func saveManifest(m Manifest) error {
	log.Printf("DEBUG: saving manifest (%d packages)", len(m.Installed))

	// ENSURE DIRECTORY EXISTS (THIS WAS MISSING)
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return err
	}

	tmp := manifestPath + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(&m); err != nil {
		f.Close()
		return err
	}
	f.Close()

	return os.Rename(tmp, manifestPath)
}

func manifestHas(name string) (*InstalledPkg, bool, error) {
	m, err := loadManifest()
	if err != nil {
		return nil, false, err
	}

	for _, p := range m.Installed {
		if p.Name == name {
			return &p, true, nil
		}
	}

	return nil, false, nil
}

func addToManifest(pkg PackageInfo) error {
	log.Printf("INFO: adding %s to manifest", pkg.Name)

	m, err := loadManifest()
	if err != nil {
		return err
	}

	for _, p := range m.Installed {
		if p.Name == pkg.Name {
			log.Printf("WARN: %s already recorded in manifest", pkg.Name)
			return nil
		}
	}

	m.Installed = append(m.Installed, InstalledPkg{
		Name:        pkg.Name,
		Version:     pkg.Version,
		Release:     int64(pkg.Release),
		InstalledAt: time.Now().Unix(),
	})

	return saveManifest(m)
}