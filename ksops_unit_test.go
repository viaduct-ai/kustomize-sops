// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/sync/errgroup"
)

func importTestKey(t *testing.T) {
	t.Helper()
	keyPath := filepath.Join("test", "key.asc")
	if _, err := os.Stat(keyPath); err != nil {
		t.Skipf("GPG test key not found at %s: %v", keyPath, err)
	}
	// Import silently; may already be imported
	cmd := exec.Command("gpg", "--batch", "--import", keyPath)
	cmd.Run()
}

func testFixturePath(t *testing.T, parts ...string) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	return abs
}

// makeManifest builds a ksops manifest YAML with absolute paths.
func makeManifest(files []string, secretFrom ...string) []byte {
	var b strings.Builder
	b.WriteString("apiVersion: viaduct.ai/v1\nkind: ksops\nmetadata:\n  name: test\n")
	if len(files) > 0 {
		b.WriteString("files:\n")
		for _, f := range files {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}
	if len(secretFrom) > 0 {
		b.WriteString(strings.Join(secretFrom, "\n"))
		b.WriteString("\n")
	}
	return []byte(b.String())
}

// --- decryptAll tests (no SOPS needed) ---

func TestDecryptAllOrderPreservation(t *testing.T) {
	var g errgroup.Group
	g.SetLimit(10)

	items := []string{"a", "b", "c", "d", "e"}
	results, err := decryptAll(&g, items, func(file string) (string, error) {
		return "result-" + file, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(results))
	}
	for i, item := range items {
		want := "result-" + item
		if results[i] != want {
			t.Errorf("results[%d] = %q, want %q", i, results[i], want)
		}
	}
}

func TestDecryptAllEmpty(t *testing.T) {
	var g errgroup.Group
	g.SetLimit(10)

	results, err := decryptAll(&g, nil, func(file string) (string, error) {
		t.Fatal("should not be called for nil input")
		return "", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}

	results, err = decryptAll(&g, []string{}, func(file string) (string, error) {
		t.Fatal("should not be called for empty input")
		return "", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestDecryptAllErrorPropagation(t *testing.T) {
	var g errgroup.Group
	g.SetLimit(10)

	sentinel := errors.New("decrypt failed")
	items := []string{"a", "b", "c"}
	_, err := decryptAll(&g, items, func(file string) (string, error) {
		if file == "b" {
			return "", sentinel
		}
		return file, nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestDecryptAllConcurrency(t *testing.T) {
	var g errgroup.Group
	g.SetLimit(5)

	var running atomic.Int32
	var maxRunning atomic.Int32

	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("file-%d", i)
	}

	results, err := decryptAll(&g, items, func(file string) (string, error) {
		cur := running.Add(1)
		// Track max concurrent goroutines
		for {
			old := maxRunning.Load()
			if cur <= old || maxRunning.CompareAndSwap(old, cur) {
				break
			}
		}
		// Small yield to let other goroutines start
		running.Add(-1)
		return "done-" + file, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(results))
	}
	for i, item := range items {
		if results[i] != "done-"+item {
			t.Errorf("results[%d] = %q, want %q", i, results[i], "done-"+item)
		}
	}
}

// --- generate tests (require GPG key) ---

func TestGenerateSingleFile(t *testing.T) {
	importTestKey(t)

	file := testFixturePath(t, "test", "legacy", "single", "secret.enc.yaml")
	manifest := makeManifest([]string{file})

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// The decrypted secret should contain these values
	for _, want := range []string{"mysecret", "username", "password", "kustomize-sops"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateMultipleFiles(t *testing.T) {
	importTestKey(t)

	dir := testFixturePath(t, "test", "legacy", "multiple")
	manifest := makeManifest([]string{
		filepath.Join(dir, "secret-A.enc.yaml"),
		filepath.Join(dir, "secret-B.enc.yaml"),
		filepath.Join(dir, "secret-C.enc.yaml"),
	})

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// Should produce 3 documents separated by ---
	docs := strings.Split(got, "---")
	if len(docs) != 3 {
		t.Errorf("expected 3 documents, got %d:\n%s", len(docs), got)
	}

	for _, name := range []string{"mysecret-A", "mysecret-B", "mysecret-C"} {
		if !strings.Contains(got, name) {
			t.Errorf("output missing secret %q:\n%s", name, got)
		}
	}
}

func TestGenerateSecretFromFiles(t *testing.T) {
	importTestKey(t)

	file := testFixturePath(t, "test", "legacy", "file", "secret.enc.yaml")
	manifest := makeManifest(nil, fmt.Sprintf(`secretFrom:
- metadata:
    name: mysecret
  type: Opaque
  files:
  - %s`, file))

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if !strings.Contains(got, "kind: Secret") {
		t.Errorf("output missing 'kind: Secret':\n%s", got)
	}
	if !strings.Contains(got, "name: mysecret") {
		t.Errorf("output missing 'name: mysecret':\n%s", got)
	}
	if !strings.Contains(got, "stringData:") {
		t.Errorf("output missing 'stringData:':\n%s", got)
	}
}

func TestGenerateSecretFromBinaryFiles(t *testing.T) {
	importTestKey(t)

	file := testFixturePath(t, "test", "legacy", "binaryfile", "secret.enc.yaml")
	manifest := makeManifest(nil, fmt.Sprintf(`secretFrom:
- metadata:
    name: mysecret
  type: Opaque
  binaryFiles:
  - %s`, file))

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if !strings.Contains(got, "data:") {
		t.Errorf("output missing 'data:':\n%s", got)
	}
	// Binary data should be base64 encoded
	if !strings.Contains(got, "name: mysecret") {
		t.Errorf("output missing 'name: mysecret':\n%s", got)
	}
}

func TestGenerateSecretFromEnvs(t *testing.T) {
	importTestKey(t)

	file := testFixturePath(t, "test", "legacy", "envs", "secret.enc.env")
	manifest := makeManifest(nil, fmt.Sprintf(`secretFrom:
- metadata:
    name: mysecret
  envs:
  - %s`, file))

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if !strings.Contains(got, "name: mysecret") {
		t.Errorf("output missing 'name: mysecret':\n%s", got)
	}
	if !strings.Contains(got, "username:") {
		t.Errorf("output missing 'username:':\n%s", got)
	}
	if !strings.Contains(got, "password:") {
		t.Errorf("output missing 'password:':\n%s", got)
	}
}

func TestGenerateFilesAndSecretFrom(t *testing.T) {
	importTestKey(t)

	dir := testFixturePath(t, "test", "legacy", "multiple")
	file := filepath.Join(dir, "secret.enc.yaml")
	manifest := makeManifest(
		[]string{
			filepath.Join(dir, "secret-A.enc.yaml"),
		},
		fmt.Sprintf(`secretFrom:
- metadata:
    name: mysecret
  type: Opaque
  binaryFiles:
  - %s`, file),
	)

	got, err := generate(manifest)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// Should have both: the decrypted file and the constructed secret
	docs := strings.Split(got, "---")
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d:\n%s", len(docs), got)
	}
}

// --- concurrency limit tests ---

func TestConcurrencyLimitEnvVar(t *testing.T) {
	importTestKey(t)

	file := testFixturePath(t, "test", "legacy", "single", "secret.enc.yaml")
	manifest := makeManifest([]string{file})

	t.Run("valid limit", func(t *testing.T) {
		t.Setenv("KSOPS_CONCURRENCY_LIMIT", "5")
		_, err := generate(manifest)
		if err != nil {
			t.Fatalf("generate failed with valid limit: %v", err)
		}
	})

	t.Run("invalid non-numeric", func(t *testing.T) {
		t.Setenv("KSOPS_CONCURRENCY_LIMIT", "abc")
		_, err := generate(manifest)
		if err == nil {
			t.Fatal("expected error for non-numeric limit")
		}
		if !strings.Contains(err.Error(), "KSOPS_CONCURRENCY_LIMIT") {
			t.Errorf("error should mention KSOPS_CONCURRENCY_LIMIT: %v", err)
		}
	})

	t.Run("invalid zero", func(t *testing.T) {
		t.Setenv("KSOPS_CONCURRENCY_LIMIT", "0")
		_, err := generate(manifest)
		if err == nil {
			t.Fatal("expected error for zero limit")
		}
	})

	t.Run("invalid negative", func(t *testing.T) {
		t.Setenv("KSOPS_CONCURRENCY_LIMIT", "-1")
		_, err := generate(manifest)
		if err == nil {
			t.Fatal("expected error for negative limit")
		}
	})

	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("KSOPS_CONCURRENCY_LIMIT", "")
		_, err := generate(manifest)
		if err != nil {
			t.Fatalf("generate failed with default limit: %v", err)
		}
	})
}

func TestGenerateErrors(t *testing.T) {
	t.Run("missing files and secretFrom", func(t *testing.T) {
		manifest := []byte(`apiVersion: viaduct.ai/v1
kind: ksops
metadata:
  name: test
`)
		_, err := generate(manifest)
		if err == nil {
			t.Fatal("expected error for missing files and secretFrom")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		_, err := generate([]byte(`{invalid`))
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		manifest := makeManifest([]string{"/nonexistent/file.yaml"})
		_, err := generate(manifest)
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}
