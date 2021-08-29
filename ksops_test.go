// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func runKSOPSPluginIntegrationTest(t *testing.T, testDir string, kustomizeVersion string) {
	want, err := ioutil.ReadFile(path.Join(testDir, "want.yaml"))
	if err != nil {
		t.Fatalf("error readding expected resources file %s: %v", path.Join(testDir, "want.yaml"), err)
	}

	pluginFlag := "--enable-alpha-plugins"
	if kustomizeVersion == "v3" {
		pluginFlag = "--enable_alpha_plugins"
	}

	cmd := exec.Command("kustomize", "build", pluginFlag, testDir)
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

	// run kustomize version to validate installation
	// and get kustomize version
	cmd := exec.Command("kustomize", "version", "--short")
	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()

	if err != nil {
		t.Errorf("error running 'kustomize version': verify kustomize is installed: %v", err)
		return
	}

	if stdErr.String() != "" {
		t.Errorf("error running 'kustomize version': verify kustomize is installed: %s", stdErr.String())
		return
	}

	// assume v4 (latest at time of writing)
	kustomizeVersion := "v4"

	if strings.Contains(stdOut.String(), "v3") {
		t.Log("detected kustomize v3")
		kustomizeVersion = "v3"
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runKSOPSPluginIntegrationTest(t, tc.dir, kustomizeVersion)
		})
	}
}
