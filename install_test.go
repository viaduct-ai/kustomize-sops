// Copyright 2024 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()

	// Create a source file with known content
	srcContent := []byte("hello world")
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, srcContent, 0755); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	dst := filepath.Join(dir, "dst")
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() returned error: %v", err)
	}

	// Verify content matches
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(got) != string(srcContent) {
		t.Errorf("content mismatch: got %q, want %q", got, srcContent)
	}

	// Verify permissions are executable
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Errorf("destination file is not executable: mode=%v", info.Mode())
	}
}

func TestCopyFileSourceNotFound(t *testing.T) {
	dir := t.TempDir()
	err := copyFile(filepath.Join(dir, "nonexistent"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Error("copyFile() should return error for nonexistent source")
	}
}

func TestCopyFileDestDirNotFound(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("data"), 0755); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	err := copyFile(src, filepath.Join(dir, "nodir", "dst"))
	if err == nil {
		t.Error("copyFile() should return error when destination directory doesn't exist")
	}
}
