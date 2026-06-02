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
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckDirAndCreate_NewDir(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "nested", "dir")

	if err := checkDirAndCreate(target); err != nil {
		t.Fatalf("checkDirAndCreate failed: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory, got file")
	}
}

func TestCheckDirAndCreate_ExistingDir(t *testing.T) {
	tmp := t.TempDir()

	if err := checkDirAndCreate(tmp); err != nil {
		t.Fatalf("checkDirAndCreate failed on existing dir: %v", err)
	}
}

func TestCompareSHA256_Match(t *testing.T) {
	tmp := t.TempDir()
	content := []byte("hello world")
	fpath := filepath.Join(tmp, "testfile")

	if err := os.WriteFile(fpath, content, 0644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	ok, err := compareSHA256(expected, fpath)
	if err != nil {
		t.Fatalf("compareSHA256 error: %v", err)
	}
	if !ok {
		t.Fatal("expected SHA256 match, got mismatch")
	}
}

func TestCompareSHA256_Mismatch(t *testing.T) {
	tmp := t.TempDir()
	content := []byte("hello world")
	fpath := filepath.Join(tmp, "testfile")

	if err := os.WriteFile(fpath, content, 0644); err != nil {
		t.Fatal(err)
	}

	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"

	ok, err := compareSHA256(wrongHash, fpath)
	if err != nil {
		t.Fatalf("compareSHA256 error: %v", err)
	}
	if ok {
		t.Fatal("expected SHA256 mismatch, got match")
	}
}

func TestCompareSHA256_CaseInsensitive(t *testing.T) {
	tmp := t.TempDir()
	content := []byte("test data")
	fpath := filepath.Join(tmp, "testfile")

	if err := os.WriteFile(fpath, content, 0644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	upperHash := hex.EncodeToString(h[:])

	ok, err := compareSHA256(upperHash, fpath)
	if err != nil {
		t.Fatalf("compareSHA256 error: %v", err)
	}
	if !ok {
		t.Fatal("expected case-insensitive SHA256 match")
	}
}

func TestCompareSHA256_FileNotFound(t *testing.T) {
	_, err := compareSHA256("abc", "/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestNormalizeYesNo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"y", "yes"},
		{"Y", "yes"},
		{"yes", "yes"},
		{"YES", "yes"},
		{"n", "no"},
		{"N", "no"},
		{"no", "no"},
		{"NO", "no"},
		{"  n  ", "no"},
		{"", "yes"},
		{"anything", "yes"},
		{"sure", "yes"},
	}

	for _, tc := range tests {
		result := normalizeYesNo(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeYesNo(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
