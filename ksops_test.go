// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func runKSOPSPluginIntegrationTest(t *testing.T, testDir string, kustomizeVersion string) {
	want, err := os.ReadFile(path.Join(testDir, "want.yaml"))
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
	if !bytes.Equal(want, got) {
		t.Errorf("wanted %s. got %s", want, got)
	}
}

func TestKSOPSPluginInstallation(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{
			name: "Legacy Single Resource",
			dir:  "test/legacy/single",
		},
		{
			name: "Legacy Multiple Resources",
			dir:  "test/legacy/multiple",
		},
		{
			name: "Legacy Hash Suffix",
			dir:  "test/legacy/hash",
		},
		{
			name: "Legacy Replace Behavior",
			dir:  "test/legacy/behaviors",
		},
		{
			name: "Legacy From File",
			dir:  "test/legacy/file",
		},
		{
			name: "Legacy From Binary File",
			dir:  "test/legacy/binaryfile",
		},
		{
			name: "Legacy From Envs",
			dir:  "test/legacy/envs",
		},
		{
			name: "Legacy Override Key",
			dir:  "test/legacy/override",
		},
		{
			name: "Legacy Secret Metadata",
			dir:  "test/legacy/metadata",
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
			name: "KRM From Binary File",
			dir:  "test/krm/binaryfile",
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
		{
			name: "KRM Secret from Template",
			dir:  "test/krm/template",
		},
		{
			name: "KRM Secret from Template using File",
			dir:  "test/krm/template-file",
		},
	}

	// run kustomize version to validate installation
	// and get kustomize version
	cmd := exec.Command("kustomize", "version")
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
