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
	"os"
	"path/filepath"
	"testing"
)

func TestPostExtractDir_SingleSubdir(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "mypackage-1.0")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := postExtractDir(tmp)
	if err != nil {
		t.Fatalf("postExtractDir failed: %v", err)
	}
	if result != subdir {
		t.Errorf("expected %q, got %q", subdir, result)
	}
}

func TestPostExtractDir_MultipleEntries(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmp, "dir2"), 0755)

	result, err := postExtractDir(tmp)
	if err != nil {
		t.Fatalf("postExtractDir failed: %v", err)
	}
	if result != tmp {
		t.Errorf("expected extract root %q, got %q", tmp, result)
	}
}

func TestPostExtractDir_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("x"), 0644)

	result, err := postExtractDir(tmp)
	if err != nil {
		t.Fatalf("postExtractDir failed: %v", err)
	}
	if result != tmp {
		t.Errorf("expected extract root %q for single file, got %q", tmp, result)
	}
}

func TestPostExtractDir_EmptyDir(t *testing.T) {
	tmp := t.TempDir()

	result, err := postExtractDir(tmp)
	if err != nil {
		t.Fatalf("postExtractDir failed: %v", err)
	}
	if result != tmp {
		t.Errorf("expected extract root %q for empty dir, got %q", tmp, result)
	}
}
