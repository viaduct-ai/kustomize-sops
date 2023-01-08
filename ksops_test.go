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

	arg := []string{"build"}
	flags := []string{"--enable-alpha-plugins", "--enable-exec"}
	if kustomizeVersion == "v3" {
		flags = []string{"--enable_alpha_plugins"}
	}
	arg = append(arg, flags...)
	arg = append(arg, testDir)

	cmd := exec.Command("kustomize", arg...)
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
			name: "Single Resource",
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
		{
			name: "From File",
			dir:  "test/file",
		},
		{
			name: "From Envs",
			dir:  "test/envs",
		},
		{
			name: "Override Key",
			dir:  "test/override",
		},
		{
			name: "Secret Metadata",
			dir:  "test/metadata",
		},
		{
			name: "KRM Single Resource",
			dir:  "test/krm/single",
		},
		{
			name: "KRM Multiple Resources",
			dir:  "test/krm/multiple",
		},
		{
			name: "KRM Hash Suffix",
			dir:  "test/krm/hash",
		},
		{
			name: "KRM Replace Behavior",
			dir:  "test/krm/behaviors",
		},
		{
			name: "KRM From File",
			dir:  "test/krm/file",
		},
		{
			name: "KRM From Envs",
			dir:  "test/krm/envs",
		},
		{
			name: "KRM Override Key",
			dir:  "test/krm/override",
		},
		{
			name: "KRM Secret Metadata",
			dir:  "test/krm/metadata",
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

	// get std out and strip "ksops.v" string to simplify version matching
	versionOutput := strings.ReplaceAll(stdOut.String(), "ksops.v", "")

	if strings.Contains(versionOutput, "v3") {
		t.Log("detected kustomize v3")
		kustomizeVersion = "v3"
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runKSOPSPluginIntegrationTest(t, tc.dir, kustomizeVersion)
		})
	}
}
