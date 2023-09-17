/*
Copyright 2019 viaduct.ai
SPDX-License-Identifier: Apache-2.0

# KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

KSOPS, or kustomize-SOPS, is a kustomize plugin for SOPS encrypted resources. KSOPS can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a kustomize plugin, KSOPS allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.
*/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/joho/godotenv"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type kubernetesSecret struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   types.ObjectMeta  `json:"metadata" yaml:"metadata"`
	Type       string            `json:"type,omitempty" yaml:"type,omitempty"`
	StringData map[string]string `json:"stringData" yaml:"stringData"`
}

type secretFrom struct {
	Files    []string         `json:"files,omitempty" yaml:"files,omitempty"`
	Envs     []string         `json:"envs,omitempty" yaml:"envs,omitempty"`
	Metadata types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Type     string           `json:"type,omitempty" yaml:"type,omitempty"`
}

type ksops struct {
	Files      []string     `json:"files,omitempty" yaml:"files,omitempty"`
	SecretFrom []secretFrom `json:"secretFrom,omitempty" yaml:"secretFrom,omitempty"`
}

func help() {
	msg := `
		KSOPS is a flexible kustomize plugin for SOPS encrypted resources.
		KSOPS supports both legacy and KRM style exec kustomize functions.

		kustomize Usage:
		- kustomize build --enable-alpha-plugins --enable-exec

		Standalone Usage :
		- Legacy: ksops secret-generator.yaml
		- KRM: cat secret-generator.yaml | ksops
`
	fmt.Fprintf(os.Stderr, "%s", strings.ReplaceAll(msg, "		", ""))
	os.Exit(1)
}

// main executes KOSPS as an exec plugin
func main() {
	nargs := len(os.Args)
	if !(nargs == 1 || nargs == 2) {
		help()
	}

	// If one argument, assume KRM style
	if nargs == 1 {
		// https://stackoverflow.com/questions/22744443/check-if-there-is-something-to-read-on-stdin-in-golang
		stat, _ := os.Stdin.Stat()

		// Check the StdIn content.
		if !(stat.Mode()&os.ModeCharDevice == 0) {
			help()
		}
		err := fn.AsMain(fn.ResourceListProcessorFunc(krm))
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to generate manifests: %v", err)
			os.Exit(1)
		}
		return
	}

	// If two argument, assume legacy style

	// ignore the first file name argument
	// load the second argument, the file path
	manifest, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read in manifest: %s", os.Args[1])
		os.Exit(1)
	}

	result, err := generate(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to generate manifests: %v", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

// https://pkg.go.dev/github.com/GoogleContainerTools/kpt-functions-sdk/go/fn#hdr-KRM_Function
func krm(rl *fn.ResourceList) (bool, error) {
	var items fn.KubeObjects
	for _, manifest := range rl.Items {
		out, err := generate([]byte(manifest.String()))
		if err != nil {
			rl.LogResult(err)
			return false, err
		}

		// generate can return multiple manifests
		objs, err := fn.ParseKubeObjects([]byte(out))
		if err != nil {
			rl.LogResult(err)
			return false, err
		}

		items = append(items, objs...)
	}

	rl.Items = items

	return true, nil
}

func generate(raw []byte) (string, error) {
	var manifest ksops
	err := yaml.Unmarshal(raw, &manifest)

	if err != nil {
		return "", fmt.Errorf("error unmarshalling manifest content: %q \n%s", err, raw)
	}

	if manifest.Files == nil && manifest.SecretFrom == nil {
		return "", fmt.Errorf("missing the required 'files' or 'secretFrom' key in the ksops manifests: %s", raw)
	}

	var output bytes.Buffer

	for i, file := range manifest.Files {
		data, err := decryptFile(file)
		if err != nil {
			return "", fmt.Errorf("error decrypting file %q from manifest.Files: %w", file, err)
		}

		output.Write(data)
		// KRM treats will try parse (and fail) empty documents if there is a trailing separator
		if i < len(manifest.Files)-1 {
			output.WriteString("\n---\n")
		}
	}

	for i, secretFrom := range manifest.SecretFrom {
		stringData := make(map[string]string)

		for _, file := range secretFrom.Files {
			key, path := fileKeyPath(file)
			data, err := decryptFile(path)
			if err != nil {
				return "", fmt.Errorf("error decrypting file %q from secretFrom.Files: %w", path, err)
			}

			stringData[key] = string(data)
		}

		for _, file := range secretFrom.Envs {
			data, err := decryptFile(file)
			if err != nil {
				return "", fmt.Errorf("error decrypting file %q from secretFrom.Envs: %w", file, err)
			}

			env, err := godotenv.Unmarshal(string(data))
			if err != nil {
				return "", fmt.Errorf("error unmarshalling .env file %q: %w", file, err)
			}
			for k, v := range env {
				stringData[k] = v
			}
		}

		s := kubernetesSecret{
			APIVersion: "v1",
			Kind:       "Secret",
			Metadata:   secretFrom.Metadata,
			Type:       secretFrom.Type,
			StringData: stringData,
		}
		d, err := yaml.Marshal(&s)
		if err != nil {
			return "", fmt.Errorf("error marshalling manifest: %w", err)
		}
		output.WriteString(string(d))
		// KRM treats will try parse (and fail) empty documents if there is a trailing separator
		if i < len(manifest.SecretFrom)-1 {
			output.WriteString("---\n")
		}
	}

	return output.String(), nil
}

func decryptFile(file string) ([]byte, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", file, err)
	}

	format := formats.FormatForPath(file)
	data, err := decrypt.DataWithFormat(b, format)
	if err != nil {
		return nil, fmt.Errorf("trouble decrypting file: %w", err)
	}
	return data, nil
}

func fileKeyPath(file string) (string, string) {
	slices := strings.Split(file, "=")
	if len(slices) == 1 {
		return filepath.Base(file), file
	} else if len(slices) > 2 {
		fmt.Fprintf(os.Stderr, "invalid format in file generator %s", file)
		os.Exit(1)
	}
	return slices[0], slices[1]
}
