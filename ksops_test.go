// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"path"
	"testing"
)

func runKSOPSPluginIntegrationTest(t *testing.T, testDir string) {
	want, err := ioutil.ReadFile(path.Join(testDir, "want.yaml"))
	if err != nil {
		t.Fatalf("error readding expected resources file %s: %v", path.Join(testDir, "want.yaml"), err)
	}

	cmd := exec.Command("kustomize", "build", "--enable-alpha-plugins", testDir)
	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()

	got := out.Bytes()
	if bytes.Compare(want, got) != 0 {
		t.Errorf("wanted %s. got %s", want, got)
	}
}

func TestKSOPSPluginInstallation(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{
			name: "Simple",
			dir:  "test/single",
		},
		{
			name: "Multiple Resources",
			dir:  "test/multiple",
		},
		{
			name: "Hash Suffix",
			dir:  "test/hash",
		},
		{
			name: "Replace Behavior",
			dir:  "test/behaviors",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runKSOPSPluginIntegrationTest(t, tc.dir)
		})
	}
}
