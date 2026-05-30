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
/*
  Blink, a powerful source-based package manager. Core of ApertureOS.
  Copyright (C) 2025-2026 Aperture OS

  Licensed under the Apache 2.0 License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ensureRepo ensures all repositories exist, are updated, verified,
// and checked out in a safe, reproducible way.
func ensureRepo(force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	repos, err := LoadConfig() // defined in config.go
	if err != nil {
		return err
	}

	// Ensure base repository directory exists
	if err := os.MkdirAll(LocalRepositoryDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo dir: %v", err)
	}

	for name, repo := range repos {
		repoPath := filepath.Join(LocalRepositoryDirPath, name)

		// Clone if missing
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			if err := cloneRepo(ctx, repo.URL, repo.Ref, repoPath); err != nil {
				return fmt.Errorf("failed to clone repository %s: %v", name, err)
			}
		}

		// Fetch only – never mutate working tree before verification
		if err := fetchRepo(ctx, repoPath); err != nil {
			return fmt.Errorf("failed to fetch repository %s: %v", name, err)
		}

		// Resolve target commit
		target, err := resolveTargetCommit(repoPath, repo.Ref)
		if err != nil {
			return fmt.Errorf("failed to resolve target for %s: %v", name, err)
		}

		// Verify pinned hash (short or full)
		if repo.Hash != "" {
			if !strings.HasPrefix(target, repo.Hash) {
				return fmt.Errorf("repository %s hash mismatch (got %s)", name, target)
			}
		}

		// Verify GPG signature against the exact trusted key
		if repo.TrustedKey != "" {
			ok, err := verifyGPGCommit(ctx, repoPath, target, filepath.Join(repoPath, repo.TrustedKey))
			if err != nil || !ok {
				return fmt.Errorf("repository %s failed GPG verification: %v", name, err)
			}
		}

		// Checkout verified commit
		if force {
			if err := checkoutCommit(ctx, repoPath, target); err != nil {
				return fmt.Errorf("failed to checkout repository %s: %v", name, err)
			}
		} else {
			// Fast-forward only when not forced
			if err := fastForwardRepo(ctx, repoPath, target); err != nil {
				return fmt.Errorf("failed to fast-forward repository %s: %v", name, err)
			}
		}
	}

	return nil
}

// cloneRepo clones a repository at a specific branch (if provided)
func cloneRepo(ctx context.Context, url, ref, dest string) error {
	args := []string{"clone"} // never ever use no-checkout i spent 30 mins tryna figure out why the cloning dir was left empty
	if ref != "" {
		args = append(args, "-b", ref)
	}
	args = append(args, url, dest)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fetchRepo fetches all refs without modifying the working tree
func fetchRepo(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "fetch", "--all", "--tags", "--prune")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// resolveTargetCommit resolves the commit hash for a ref or HEAD
func resolveTargetCommit(path, ref string) (string, error) {
	refspec := "FETCH_HEAD"
	if ref != "" {
		refspec = "origin/" + ref
	}

	cmd := exec.Command("git", "-C", path, "rev-parse", refspec)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// checkoutCommit hard-resets the working tree to a verified commit
func checkoutCommit(ctx context.Context, path, commit string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "reset", "--hard", commit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fastForwardRepo fast-forwards HEAD to the target commit if possible
func fastForwardRepo(ctx context.Context, path, commit string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "merge", "--ff-only", commit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// verifyGPGCommit verifies that a specific commit is signed by the trusted key
// Uses an isolated GNUPGHOME and enforces fingerprint matching
func verifyGPGCommit(ctx context.Context, repoPath, commit, pubKeyPath string) (bool, error) {
	gnupgHome, err := os.MkdirTemp("", "blink-gnupg-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(gnupgHome)

	env := append(os.Environ(), "GNUPGHOME="+gnupgHome)

	// Import trusted key
	cmd := exec.CommandContext(ctx, "gpg", "--import", pubKeyPath)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to import key: %v\n%s", err, string(out))
	}

	// Extract expected fingerprint
	cmd = exec.CommandContext(ctx, "gpg", "--with-colons", "--fingerprint")
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	var expectedFP string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "fpr:") {
			parts := strings.Split(line, ":")
			expectedFP = parts[9]
			break
		}
	}
	if expectedFP == "" {
		return false, fmt.Errorf("could not extract trusted key fingerprint")
	}

	// Verify commit and capture signer fingerprint
	cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "verify-commit", "--raw", commit)
	cmd.Env = env
	out, err = cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("GPG verification failed: %v\n%s", err, string(out))
	}

	if !strings.Contains(string(out), expectedFP) {
		return false, fmt.Errorf("commit not signed by trusted key")
	}

	return true, nil
}

var reposEnsured bool

func ensureRepoOnce(force bool) error {
	if reposEnsured {
		return nil
	}
	reposEnsured = true
	return ensureRepo(force)
}

// FindRepoForPackage returns the repo and recipe path for a package
// FindRepoForPackage searches all configured repositories for the given package name.
// Returns the repository that contains the package and the full path to the package JSON.
func FindRepoForPackage(pkgName string, repos map[string]RepoConfig) (RepoConfig, string, error) {
	type match struct {
		repo RepoConfig
		path string
	}

	var matches []match

	for _, repo := range repos {
		recipePath := filepath.Join(
			LocalRepositoryDirPath,
			repo.Name,
			"recipes",
			pkgName+".json", // package JSON file
		)

		if _, err := os.Stat(recipePath); err == nil {
			matches = append(matches, match{
				repo: repo,
				path: recipePath,
			})
		}
	}

	switch len(matches) {
	case 0:
		return RepoConfig{}, "", fmt.Errorf(
			"package %q not found in any configured repository",
			pkgName,
		)

	case 1:
		return matches[0].repo, matches[0].path, nil

	default:
		fmt.Printf("Multiple repositories provide package %q:\n\n", pkgName)
		for i, m := range matches {
			fmt.Printf(" [%d] %s (%s)\n", i+1, m.repo.Name, m.repo.URL)
		}

		fmt.Print("\nSelect repository number: ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(matches) {
			return RepoConfig{}, "", fmt.Errorf("invalid selection")
		}

		selected := matches[choice-1]
		return selected.repo, selected.path, nil
	}
}
