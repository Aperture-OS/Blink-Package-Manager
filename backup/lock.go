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
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/Aperture-OS/eyes"
)

// Lock represents a file-based lock for a given path.
// This allows multiple independent locks without a global variable.
type Lock struct {
	Path string   // Path to the lock file
	file *os.File // Internal file handle used for locking
}

// Acquire tries to acquire an exclusive lock on the lock file (non-blocking).
// Returns an error if another process is already holding the lock.
func (l *Lock) Acquire() error {
	f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_RDWR, 0600) // safer perms
	if err != nil {
		return fmt.Errorf("failed to open lock file: %v", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			return fmt.Errorf("another instance is running")
		}
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	if err := f.Truncate(0); err != nil {
		f.Close()
		return fmt.Errorf("failed to truncate lock file: %v", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return fmt.Errorf("failed to seek lock file: %v", err)
	}

	if _, err := fmt.Fprintf(f, "%d\n", os.Getpid()); err != nil {
		f.Close()
		return fmt.Errorf("failed to write PID: %v", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("failed to sync lock file: %v", err)
	}

	l.file = f
	eyes.Infof("Lock acquired at %s", l.Path)
	return nil
}

// Release releases the lock and closes the file.
func (l *Lock) Release() error {
	if l.file == nil {
		return fmt.Errorf("lock not acquired, cannot release")
	}

	// Release the file lock
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %v", err)
	}

	// Close the file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %v", err)
	}

	eyes.Infof("Lock released at %s", l.Path)
	l.file = nil
	return nil
}

// IsLocked checks whether the lock file is currently locked by another process.
// It attempts to acquire the lock non-blocking and immediately releases it.
// Note: This check is advisory and may race with other processes.
func (l *Lock) IsLocked() (bool, error) {
	// Open (or create) the lock file
	f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open lock file: %v", err)
	}
	defer f.Close()

	// Try to acquire exclusive lock non-blocking
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		// Check both EAGAIN and EWOULDBLOCK
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			return true, nil // Locked by another process
		}
		return false, fmt.Errorf("error checking lock: %v", err)
	}

	// Lock acquired successfully, immediately release it
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		return false, fmt.Errorf("failed to unlock after check: %v", err)
	}

	return false, nil
}
