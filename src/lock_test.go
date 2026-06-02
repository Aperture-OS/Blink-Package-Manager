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

func TestLock_AcquireAndRelease(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	l := &Lock{Path: lockPath}

	if err := l.Acquire(); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if l.file == nil {
		t.Fatal("expected file handle to be set after Acquire")
	}

	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatal("lock file should exist after Acquire")
	}

	if err := l.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	if l.file != nil {
		t.Fatal("expected file handle to be nil after Release")
	}
}

func TestLock_DoubleRelease(t *testing.T) {
	l := &Lock{Path: "/tmp/unused.lock"}

	err := l.Release()
	if err == nil {
		t.Fatal("expected error when releasing without acquiring")
	}
}

func TestLock_IsLocked_WhenFree(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	l := &Lock{Path: lockPath}

	locked, err := l.IsLocked()
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if locked {
		t.Fatal("expected lock to be free")
	}
}

func TestLock_IsLocked_WhenHeld(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	holder := &Lock{Path: lockPath}
	if err := holder.Acquire(); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	defer holder.Release()

	checker := &Lock{Path: lockPath}
	locked, err := checker.IsLocked()
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if !locked {
		t.Fatal("expected lock to be held")
	}
}

func TestLock_Acquire_BlockedByAnother(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	first := &Lock{Path: lockPath}
	if err := first.Acquire(); err != nil {
		t.Fatalf("first Acquire failed: %v", err)
	}
	defer first.Release()

	second := &Lock{Path: lockPath}
	err := second.Acquire()
	if err == nil {
		second.Release()
		t.Fatal("expected second Acquire to fail")
	}
}

func TestLock_WritesPID(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	l := &Lock{Path: lockPath}
	if err := l.Acquire(); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	defer l.Release()

	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("lock file should contain PID")
	}
}
