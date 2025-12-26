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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

/****************************************************/
// getPkg downloads a package recipe from the repository and saves it to the specified path
// you can use this standalone to just download recipes if you want, but usually this is
// called internally by other functions, ensuring reusing code, modularity, and less repetition
/****************************************************/

func getpkg(pkgName string, path string) error {

	log.Printf("INFO: Getting package recipe.")

	log.Printf("INFO: Acquiring lock at %s", lockPath)
	if checkLock(lockPath) {
		return fmt.Errorf("another instance is running, lock file exists at %s", lockPath)
	}

	lockErr := addLock(lockPath) // add lock file
	defer removeLock(lockPath)   // remove lock file at the end

	if lockErr != nil { // check for errors while adding lock
		return fmt.Errorf("failed to create lock file at %s: %v", lockPath, lockErr)
	}

	// Ensure path ends with OS-specific separator
	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}

	// Check if cache directory exists

	checkDirAndCreate(path)
	checkDirAndCreate(filepath.Join(path, "recipes"))

	// Full path to recipe
	recipePath := filepath.Join(path, "recipes", pkgName+".json")

	// Build URL for package JSON
	url := baseURL + pkgName + ".json"

	// Perform HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download recipe: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download recipe, status: %s", resp.Status)
	}

	// Create file to save recipe
	outFile, err := os.Create(recipePath)
	if err != nil {
		return fmt.Errorf("failed to create recipe file: %v", err)
	}
	defer outFile.Close()

	// Copy response body to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write recipe file: %v", err)
	}

	log.Printf("INFO: Package recipe downloaded to %s", recipePath)

	return nil
}

/********************************************************************************************************/

/****************************************************/
// fetchPkg fetches a package recipe from cache or repository, decodes it, and displays package info
// in addition, it returns the PackageInfo struct for further use, so you can use this function to both
// get the struct and show the info to the user, avoiding code repetition and enhancing modularity
// avoids 2 functions for fetching and displaying info separately
/****************************************************/

func fetchpkg(path string, force bool, pkgName string) (PackageInfo, error) {

	log.Printf("INFO: Fetching package %q", pkgName)

	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}

	recipePath := filepath.Join(path, "recipes", pkgName+".json")

	if force {
		if err := os.Remove(recipePath); err == nil {
			log.Printf("INFO: Force flag detected, removed cached recipe at %s", recipePath)
		} else if !os.IsNotExist(err) {
			log.Printf("WARNING: Failed to remove cached recipe.\nERR: %v", err)
		}
	}

	if _, err := os.Stat(recipePath); os.IsNotExist(err) {
		log.Printf("INFO: Package recipe not found. Downloading...")
		if err := getpkg(pkgName, path); err != nil {
			return PackageInfo{}, err
		}
	}

	f, err := os.Open(recipePath)
	if err != nil {
		log.Printf("FATAL: Failed to open package recipe.\nERR: %v", err)
		return PackageInfo{}, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	var pkg PackageInfo
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		log.Printf("FATAL: Failed to parse JSON.\nERR: %v", err)
		return PackageInfo{}, fmt.Errorf("error decoding JSON: %v", err)
	}

	fmt.Printf("\nRepository: %s\n\nName: %s\nVersion: %s\nDescription: %s\nAuthor: %s\nLicense: %s\n",
		repoURL, pkg.Name, pkg.Version, pkg.Description, pkg.Author, pkg.License)

	log.Printf("INFO: Package fetching completed.")

	return pkg, nil
}

/********************************************************************************************************/

/****************************************************/



/****************************************************/

/****************************************************/
// install function downloads, decompresses, builds, and installs a package
// it fetches package info, downloads source, decompresses it
// it uses the getSource, decompressSource functions for modularity and to satisfy my KISS principle
// i wish golang had macros so i could avoid writing the same error handling code every single time and just have a single line for it
/****************************************************/

func install(pkgName string, force bool, path string) error {
	log.Printf("INFO: ===== INSTALL START =====")
	log.Printf("INFO: pkg=%s force=%v", pkgName, force)

	// manifest must exist BEFORE touching it
	if err := ensureManifest(); err != nil {
		return err
	}

	// fetch recipe
	pkg, err := fetchpkg(path, force, pkgName)
	if err != nil {
		return err
	}

	installed, exists, err := manifestHas(pkg.Name)
	if err != nil {
		return err
	}

	if exists && !force {
		return fmt.Errorf(
			"package %s already installed (version=%s release=%d)",
			installed.Name,
			installed.Version,
			installed.Release,
		)
	}

	// prepare build root
	if err := os.MkdirAll(buildRoot, 0755); err != nil {
		return err
	}

	extractRoot := filepath.Join(buildRoot, pkg.Name)
	log.Printf("INFO: extract root = %s", extractRoot)

	_ = os.RemoveAll(extractRoot)
	if err := os.MkdirAll(extractRoot, 0755); err != nil {
		return err
	}

	// download source
	if err := getSource(pkg.Source.URL, force); err != nil {
		return err
	}

	srcFile := filepath.Join(sourcePath, filepath.Base(pkg.Source.URL))
	ok, err := compareSHA256(pkg.Source.Sha256, srcFile)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("source hash mismatch for %s", srcFile)
	}

	// extract
	if err := decompressSource(pkg, extractRoot); err != nil {
		return err
	}

	buildDir, err := postExtractDir(extractRoot)
	if err != nil {
		return err
	}

	log.Printf("INFO: build dir = %s", buildDir)

	if err := os.Chdir(buildDir); err != nil {
		return err
	}

	// env
	for k, v := range pkg.Build.Env {
		log.Printf("DEBUG: env %s=%s", k, v)
		os.Setenv(k, v)
	}

	// prepare
	for _, cmd := range pkg.Build.Prepare {
		log.Printf("INFO: prepare → %s", cmd)
		if err := runCmd("sh", "-c", cmd); err != nil {
			return err
		}
	}

	// install
	for _, cmd := range pkg.Build.Install {
		log.Printf("INFO: install → %s", cmd)
		if err := runCmd("sh", "-c", cmd); err != nil {
			return err
		}
	}

	// record install (THIS was missing/broken before)
	if err := addToManifest(pkg); err != nil {
		return err
	}

	log.Printf("INFO: ===== INSTALL COMPLETE =====")
	return nil
}

/*
 *  i wonder what "finding hidden gems in blink source code" would feel like lmao
 *  well just know that this code is open source, so feel free to explore it and find any hidden gems
 *  or pull request and add a couple ;) (no mister "i wanna contribute to foss", this doesnt count as a
 *  proper contribution but if u add gems good job!)
 */

/********************************************************************************************************/

/****************************************************/
// clean cleans the cache folders, yes thats it
/****************************************************/

func clean() error {

	fmt.Printf("WARNING: Are you sure you want to delete the cached recipes and sources? [ (Y)es / (N)o ] ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(response)
	response = strings.TrimSpace(response)

	switch response {

	case "y", "yes", "sure", "yep", "ye", "yea", "yeah", "", " ", "  ", "   ", "\n":
		log.Printf("INFO: Acquiring lock at %s", lockPath)
		if checkLock(lockPath) {
			return fmt.Errorf("another instance is running, lock file exists at %s", lockPath)
		}

		lockErr := addLock(lockPath) // add lock file
		defer removeLock(lockPath)   // remove lock file at the end

		if lockErr != nil { // check for errors while adding lock
			return fmt.Errorf("failed to create lock file at %s: %v", lockPath, lockErr)
		}

		os.RemoveAll(recipePath)
		os.MkdirAll(recipePath, 0755)

		os.RemoveAll(sourcePath)
		os.MkdirAll(sourcePath, 0755)

		os.RemoveAll(buildRoot)
		os.MkdirAll(buildRoot, 0755)

	default:
		log.Fatalf("\nFATAL: User declined, exiting...")

	}

	return nil

}