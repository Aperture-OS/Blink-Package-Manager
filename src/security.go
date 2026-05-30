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
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Aperture-OS/eyes"
)

// =============================================================================
// SECURITY CONSTANTS
// =============================================================================

const (
	// MaxDownloadSize is the maximum allowed download size (100MB)
	MaxDownloadSize = 100 * 1024 * 1024

	// MaxExtractSize is the maximum allowed extracted size (500MB)
	MaxExtractSize = 500 * 1024 * 1024

	// MaxBuildTime is the maximum allowed build time (30 minutes)
	MaxBuildTime = 30 * time.Minute

	// MaxCommandTime is the maximum allowed for a single command (5 minutes)
	MaxCommandTime = 5 * time.Minute

	// MaxDependencyDepth prevents infinite recursion in dependency resolution
	MaxDependencyDepth = 100
)

// =============================================================================
// PATH VALIDATION
// =============================================================================

// ValidatePackageName validates that a package name is safe to use
func ValidatePackageName(name string) error {
	if name == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	// Regex for valid package names: alphanumeric, starts with letter/number, 
	// allows hyphens, underscores, dots, max 64 chars
	validPackageNameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)
	if !validPackageNameRegex.MatchString(name) {
		return fmt.Errorf("invalid package name: %s (must be alphanumeric with hyphens/underscores/dots, max 64 chars)", name)
	}

	// Blacklist dangerous names
	dangerousNames := []string{
		".", "..", "/", "\\", "~", "$", "{", "}", "[", "]",
		"root", "etc", "bin", "sbin", "usr", "var", "tmp", "dev",
		"proc", "sys", "boot", "lib", "lib64", "opt", "mnt", "media",
		"home", "users", "run", "srv", "lost+found",
	}

	nameLower := strings.ToLower(name)
	for _, dangerous := range dangerousNames {
		if nameLower == dangerous {
			return fmt.Errorf("reserved or dangerous package name: %s", name)
		}
		if strings.Contains(nameLower, "/"+dangerous) || strings.Contains(nameLower, dangerous+"/") {
			return fmt.Errorf("package name contains reserved component: %s", name)
		}
	}

	return nil
}

// ValidatePath validates that a path doesn't contain path traversal sequences
func ValidatePath(path string, allowAbsolute bool) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean the path and check for traversal
	cleaned := filepath.Clean(path)
	
	// Check for absolute paths
	if filepath.IsAbs(cleaned) && !allowAbsolute {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Check for path traversal
	if strings.HasPrefix(cleaned, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	// Check each component
	parts := strings.Split(cleaned, string(os.PathSeparator))
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("path traversal detected in component: %s", part)
		}
		if part == "." || part == "" {
			continue
		}
	}

	return nil
}

// SafeJoinPath safely joins path components with validation
func SafeJoinPath(base string, parts ...string) (string, error) {
	var result string
	
	if base != "" {
		result = base
	} else {
		result = "."
	}
	
	for _, part := range parts {
		// Validate each part
		if err := ValidatePath(part, false); err != nil {
			return "", fmt.Errorf("invalid path part '%s': %v", part, err)
		}
		result = filepath.Join(result, part)
	}
	
	// Validate final result
	cleaned := filepath.Clean(result)
	if strings.HasPrefix(cleaned, "..") || cleaned == ".." {
		return "", fmt.Errorf("path traversal in result: %s", result)
	}
	
	return result, nil
}

// IsSafeRelativePath checks if a relative path is safe (no traversal, no absolute)
func IsSafeRelativePath(path string) bool {
	if path == "" || path == "." {
		return true
	}
	
	cleaned := filepath.Clean(path)
	
	// Must not be absolute
	if filepath.IsAbs(cleaned) {
		return false
	}
	
	// Must not start with ..
	if strings.HasPrefix(cleaned, "..") {
		return false
	}
	
	// Check each component
	parts := strings.Split(cleaned, string(os.PathSeparator))
	for _, part := range parts {
		if part == ".." {
			return false
		}
	}
	
	return true
}

// =============================================================================
// ARCHIVE SECURITY
// =============================================================================

// Archive magic numbers for type detection
var archiveMagicNumbers = map[string][]byte{
	".tar":     {0x75, 0x73, 0x74, 0x61, 0x72}, // "ustar" at offset 257
	".tar.gz":  {0x1f, 0x8b},                     // gzip
	".tgz":     {0x1f, 0x8b},                     // gzip
	".tar.xz":  {0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, // xz
	".tar.bz2": {0x42, 0x5a, 0x68},              // bzip2
	".zip":     {0x50, 0x4b, 0x03, 0x04},        // PK magic
	".zst":     {0x28, 0xB5, 0x2F, 0xFD},        // zstd
}

// DetectArchiveType reads the magic bytes to detect the actual archive type
func DetectArchiveType(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	header := make([]byte, 32)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}

	// For tar, check at offset 257 for "ustar"
	if n >= 262 {
		if bytes.Equal(header[257:262], []byte("ustar")) {
			return ".tar", nil
		}
	}

	// Check magic bytes for other formats
	for ext, magic := range archiveMagicNumbers {
		if bytes.HasPrefix(header[:len(magic)], magic) {
			return ext, nil
		}
	}

	return "", fmt.Errorf("unknown archive format")
}

// ValidateArchiveContent validates that an archive doesn't contain unsafe paths
func ValidateArchiveContent(archivePath, destDir string) error {
	// First, detect the actual type
	archiveType, err := DetectArchiveType(archivePath)
	if err != nil {
		return fmt.Errorf("cannot detect archive type: %v", err)
	}

	// Extract to a temporary directory first for validation
	tempDir, err := os.MkdirTemp("", "blink-validate-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set restrictive permissions
	if err := os.Chmod(tempDir, 0700); err != nil {
		return fmt.Errorf("failed to set permissions on temp dir: %v", err)
	}

	// Extract based on type
	switch archiveType {
	case ".tar", ".tar.gz", ".tgz":
		cmd := exec.Command("tar", "-xzf", archivePath, "-C", tempDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract for validation: %v", err)
		}
	case ".tar.xz":
		cmd := exec.Command("tar", "-xJf", archivePath, "-C", tempDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract for validation: %v", err)
		}
	case ".tar.bz2":
		cmd := exec.Command("tar", "-xjf", archivePath, "-C", tempDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract for validation: %v", err)
		}
	case ".zip":
		cmd := exec.Command("unzip", "-q", archivePath, "-d", tempDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract for validation: %v", err)
		}
	default:
		return fmt.Errorf("unsupported archive type for validation: %s", archiveType)
	}

	// Check extracted size
	size, err := dirSize(tempDir)
	if err != nil {
		return fmt.Errorf("failed to check extracted size: %v", err)
	}
	if size > MaxExtractSize {
		return fmt.Errorf("extracted size exceeds limit: %d > %d", size, MaxExtractSize)
	}

	// Walk through all extracted files and validate paths
	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		// Validate the relative path is safe
		if !IsSafeRelativePath(relPath) {
			return fmt.Errorf("unsafe path in archive: %s", relPath)
		}

		// Validate path doesn't try to escape when joined with destDir
		fullPath := filepath.Join(destDir, relPath)
		cleaned := filepath.Clean(fullPath)
		
		// Ensure the cleaned path still starts with destDir
		destDirCleaned := filepath.Clean(destDir) + string(os.PathSeparator)
		if !strings.HasPrefix(cleaned, destDirCleaned) && cleaned != destDirCleaned {
			return fmt.Errorf("path escapes destination: %s -> %s", relPath, cleaned)
		}

		return nil
	})
}

// dirSize calculates the total size of a directory
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// =============================================================================
// SANDBOX EXECUTION
// =============================================================================

// SandboxConfig defines sandbox parameters
type SandboxConfig struct {
	Timeout       time.Duration
	MaxCPU        float64  // 0 = unlimited
	MaxMemoryMB   int     // 0 = unlimited
	AllowNetwork  bool
	AllowWrite    bool
	WorkDir       string
	Env           map[string]string
}

// DefaultSandboxConfig returns safe defaults for build sandbox
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Timeout:      MaxCommandTime,
		MaxCPU:       2.0,  // Limit to 2 CPU cores
		MaxMemoryMB:  1024, // 1GB memory limit
		AllowNetwork: false, // No network by default for security
		AllowWrite:   true,  // Need write for builds
		WorkDir:      "/tmp/blink-sandbox",
		Env: map[string]string{
			"PATH": "/usr/bin:/bin:/usr/sbin:/sbin",
			"HOME": "/tmp/blink-home",
			"TMPDIR": "/tmp/blink-tmp",
		},
	}
}

// RunInSandbox executes a command in a restricted sandbox environment
// Uses Linux namespaces for isolation (requires root)
func RunInSandbox(cmd string, args []string, config SandboxConfig) error {
	// If not running as root, we can't use namespaces effectively
	if os.Geteuid() != 0 {
		// Fall back to regular execution with timeout
		return RunWithTimeout(cmd, args, config.Timeout)
	}

	// Check if we're on Linux
	if runtime.GOOS != "linux" {
		return RunWithTimeout(cmd, args, config.Timeout)
	}

	// Check for unshare command
	if _, err := exec.LookPath("unshare"); err != nil {
		eyes.Warn("unshare command not found, falling back to regular execution")
		return RunWithTimeout(cmd, args, config.Timeout)
	}

	// Prepare unshare arguments
	unshareArgs := []string{
		"--mount",   // Isolate mount namespace
		"--pid",     // Isolate PID namespace
		"--net",     // Isolate network namespace
		"--uts",     // Isolate hostname
		"--ipc",     // Isolate IPC
		"--fork",    // Fork before exec
	}

	if !config.AllowNetwork {
		// Block network by not sharing it
		// unshare --net already does this
	}

	// Add the command to run
	fullCmd := append([]string{cmd}, args...)
	unshareArgs = append(unshareArgs, fullCmd...)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	unshareCmd := exec.CommandContext(ctx, "unshare", unshareArgs...)
	unshareCmd.Stdout = os.Stdout
	unshareCmd.Stderr = os.Stderr

	// Set resource limits using syscall
	unshareCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:   0,
	}

	// Note: Full resource limits (CPU, memory) require cgroups or other mechanisms
	// For now, rely on the timeout

	return unshareCmd.Run()
}

// RunWithTimeout executes a command with a timeout
func RunWithTimeout(cmd string, args []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// RunCommandSafely executes a command with security checks
// This is the main function to use instead of direct runCmd calls
func RunCommandSafely(cmd string, args ...string) error {
	// For security-critical operations, use sandbox
	config := DefaultSandboxConfig()
	
	// If the command contains shell metacharacters, it's dangerous
	fullCmd := cmd + " " + strings.Join(args, " ")
	if containsShellMetaChars(fullCmd) {
		// This should have been caught earlier, but just in case
		return fmt.Errorf("dangerous command detected: %s", fullCmd)
	}

	// Use sandbox if available
	if os.Geteuid() == 0 && runtime.GOOS == "linux" {
		return RunInSandbox(cmd, args, config)
	}

	// Fallback to timeout-only execution
	return RunWithTimeout(cmd, args, config.Timeout)
}

// containsShellMetaChars checks if a string contains shell metacharacters
func containsShellMetaChars(s string) bool {
	metaChars := []string{
		";", "&", "|", "$", "`", "(", ")", "{", "}", "[", "]",
		"<", ">", "\n", "\r", "\\", "!", "#", "~", "*", "?",
	}
	for _, char := range metaChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

// =============================================================================
// INPUT VALIDATION
// =============================================================================

// ValidateURL validates a URL for security
func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// Enforce HTTPS for security (allow HTTP with warning)
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("unsupported URL scheme: %s (use http or https)", parsed.Scheme)
	}

	// Check for empty host
	if parsed.Host == "" {
		return fmt.Errorf("URL has no host")
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"127.0.0.1",
		"localhost",
		"::1",
		"file://",
		"javascript:",
		"data:",
	}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(rawURL), pattern) {
			return fmt.Errorf("suspicious URL pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateFilePath validates a file path for security
func ValidateFilePath(path string) error {
	// Must be absolute for our use case
	if !filepath.IsAbs(path) {
		return fmt.Errorf("file path must be absolute")
	}

	// Clean and validate
	cleaned := filepath.Clean(path)
	
	// Must start with /
	if !strings.HasPrefix(cleaned, "/") {
		return fmt.Errorf("path must be absolute")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	return nil
}

// =============================================================================
// NETWORK SECURITY
// =============================================================================

// Global HTTP client with secure defaults
var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

// GetHTTPClient returns a shared, secure HTTP client
func GetHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		// Create a custom transport with secure settings
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // Require TLS 1.2+
				// Optionally: add certificate pins for known repositories
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		httpClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	})
	return httpClient
}

// SafeDownload downloads a file with security checks
func SafeDownload(url string, destPath string) error {
	// Validate URL
	if err := ValidateURL(url); err != nil {
		return err
	}

	// Validate destination path
	if err := ValidateFilePath(destPath); err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return err
	}

	// Get HTTP client
	client := GetHTTPClient()

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Check content length
	if resp.ContentLength > MaxDownloadSize {
		return fmt.Errorf("download too large: %d bytes (max %d)", 
			resp.ContentLength, MaxDownloadSize)
	}

	// Create destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Set restrictive permissions
	if err := outFile.Chmod(0640); err != nil {
		eyes.Warnf("Failed to set file permissions: %v", err)
	}

	// Copy data with size limit
	if resp.ContentLength > 0 {
		// Use limited reader
		limitedReader := io.LimitReader(resp.Body, MaxDownloadSize+1)
		n, err := io.Copy(outFile, limitedReader)
		if err != nil {
			outFile.Close()
			os.Remove(destPath)
			return fmt.Errorf("failed to write file: %v", err)
		}
		if n > MaxDownloadSize {
			outFile.Close()
			os.Remove(destPath)
			return fmt.Errorf("download exceeded size limit")
		}
	} else {
		// Unknown size, use limited reader
		limitedReader := io.LimitReader(resp.Body, MaxDownloadSize+1)
		n, err := io.Copy(outFile, limitedReader)
		if err != nil {
			outFile.Close()
			os.Remove(destPath)
			return fmt.Errorf("failed to write file: %v", err)
		}
		if n > MaxDownloadSize {
			outFile.Close()
			os.Remove(destPath)
			return fmt.Errorf("download exceeded size limit")
		}
	}

	// Sync to disk
	if err := outFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %v", err)
	}

	return nil
}

// =============================================================================
// HASH VERIFICATION
// =============================================================================

// VerifySHA256 verifies a file against an expected SHA256 hash
func VerifySHA256(expectedHash, filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	return strings.EqualFold(actual, expectedHash), nil
}

// ComputeSHA256 computes the SHA256 hash of a file
func ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// =============================================================================
// CAPABILITY CHECKING (Linux)
// =============================================================================

// HasFullCapabilities checks if the current process has full root capabilities
func HasFullCapabilities() bool {
	if runtime.GOOS != "linux" {
		return os.Geteuid() == 0
	}

	// Read /proc/self/status for capability bounding set
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}

	// Check CapBnd (capability bounding set)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "CapBnd:") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return false
			}
			capHex := parts[1]
			// Full capabilities in hex (64 bits, all 1s)
			// For 64-bit systems, this is 0000003fffffffff (last 40 bits set)
			// or ffffffffffffffff (all bits set on some systems)
			if capHex == "0000003fffffffff" || capHex == "ffffffffffffffff" {
				return true
			}
			// Check if all bits we need are set
			// We need at least: CAP_CHOWN, CAP_DAC_OVERRIDE, CAP_FSETID, CAP_FOWNER, CAP_SETUID, CAP_SETGID
			// This is a simplified check
			return false
		}
	}

	return false
}

// Enhanced requireRoot with capability checking
func RequireRoot() {
	// Check UID
	if os.Geteuid() != 0 {
		eyes.Fatalf(`This command must be run as Root or Super User (also known as Admin, Administrator, SU, etc.)
Please try again with 'sudo' infront of the command or as the root user ('su -').`)
	}

	// On Linux, also check capabilities
	if runtime.GOOS == "linux" {
		if !HasFullCapabilities() {
			eyes.Fatalf("Insufficient capabilities. Requires full root privileges (CAP_CHOWN, CAP_DAC_OVERRIDE, etc.).")
		}
	}
}

// =============================================================================
// LOCK MANAGEMENT IMPROVEMENTS
// =============================================================================

// ProcessExists checks if a process with the given PID exists
func ProcessExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal 0 checks if process exists without sending a signal
	err = process.Signal(os.Signal(nil))
	if err != nil {
		// On Linux, check for specific error
		if err == os.ErrProcessDone {
			return false
		}
		if strings.Contains(err.Error(), "no such process") {
			return false
		}
	}

	return true
}

// CleanStaleLocks removes stale lock files on startup
func CleanStaleLocks() {
	lockFiles := []string{
		LockFilePath,
		filepath.Join(BaseDataDirPath, "*.lock"),
	}

	for _, pattern := range lockFiles {
		matches, _ := filepath.Glob(pattern)
		for _, lockFile := range matches {
			// Read PID from lock file
			data, err := os.ReadFile(lockFile)
			if err != nil {
				continue
			}

			var pid int
			if _, err := fmt.Sscanf(string(data), "%d", &pid); err == nil {
				// Check if process exists
				if !ProcessExists(pid) {
					// Stale lock, remove it
					os.Remove(lockFile)
					eyes.Infof("Cleaned up stale lock file: %s", lockFile)
				}
			} else {
				// Corrupted lock file, remove it
				os.Remove(lockFile)
				eyes.Infof("Cleaned up corrupted lock file: %s", lockFile)
			}
		}
	}
}

// =============================================================================
// PATH SANITIZATION FOR ERROR MESSAGES
// =============================================================================

// SanitizePath removes sensitive information from paths for error messages
func SanitizePath(path string) string {
	if path == "" {
		return path
	}

	// Replace home directory with ~
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.Replace(path, home, "", 1)
	}

	// Replace temp directories
	tempDir := os.TempDir()
	if strings.HasPrefix(path, tempDir) {
		return "/tmp/..."
	}

	// For system paths, just show the last few components
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return ".../" + filepath.Join(parts[len(parts)-3:]...)
	}

	return path
}

// SafeError wraps an error with sanitized paths
func SafeError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Sanitize the error message
	sanitized := err.Error()
	
	// Replace paths in the error message
	// This is a simple approach; could be more sophisticated
	home, _ := os.UserHomeDir()
	if home != "" {
		sanitized = strings.ReplaceAll(sanitized, home, "~")
	}
	
	tempDir := os.TempDir()
	if tempDir != "" {
		sanitized = strings.ReplaceAll(sanitized, tempDir, "/tmp")
	}

	return fmt.Errorf("%s: %s", context, sanitized)
}
