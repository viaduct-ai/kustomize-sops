/*
Copyright 2019 viaduct.ai
SPDX-License-Identifier: Apache-2.0

# KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

KSOPS, or kustomize-SOPS, is a kustomize plugin for SOPS encrypted resources. KSOPS can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a kustomize plugin, KSOPS allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.
*/
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type keyData struct {
	key  string
	data []byte
}

// decryptAll concurrently decrypts a list of files using the provided errgroup,
// returning results in the same order as the input.
func decryptAll[T any](g *errgroup.Group, items []string, fn func(file string) (T, error)) ([]T, error) {
	results := make([]T, len(items))
	for i, file := range items {
		g.Go(func() error {
			v, err := fn(file)
			if err != nil {
				return err
			}
			results[i] = v
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

type kubernetesSecret struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   types.ObjectMeta  `json:"metadata" yaml:"metadata"`
	Type       string            `json:"type,omitempty" yaml:"type,omitempty"`
	StringData map[string]string `json:"stringData,omitempty" yaml:"stringData,omitempty"`
	Data       map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

type secretFrom struct {
	Files       []string         `json:"files,omitempty" yaml:"files,omitempty"`
	BinaryFiles []string         `json:"binaryFiles,omitempty" yaml:"binaryFiles,omitempty"`
	Envs        []string         `json:"envs,omitempty" yaml:"envs,omitempty"`
	Metadata    types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Type        string           `json:"type,omitempty" yaml:"type,omitempty"`
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

// main executes KSOPS as an exec plugin
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
			fmt.Fprintf(os.Stderr, "unable to generate manifests: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// If two argument, assume legacy style

	// ignore the first file name argument
	// load the second argument, the file path
	manifest, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read in manifest: %s\n", os.Args[1])
		os.Exit(1)
	}

	result, err := generate(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to generate manifests: %v\n", err)
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

	var g errgroup.Group
	limit := 20
	if l := os.Getenv("KSOPS_CONCURRENCY_LIMIT"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 {
			return "", fmt.Errorf("error parsing KSOPS_CONCURRENCY_LIMIT value %q: %w", l, err)
		}
	}
	g.SetLimit(limit)

	// Decrypt manifest.Files concurrently
	decrypted, err := decryptAll(&g, manifest.Files, func(file string) ([]byte, error) {
		data, err := decryptFile(file)
		if err != nil {
			return nil, fmt.Errorf("error decrypting file %q from manifest.Files: %w", file, err)
		}
		return data, nil
	})
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	for i, data := range decrypted {
		output.Write(data)
		// KRM treats will try parse (and fail) empty documents if there is a trailing separator
		if i < (len(manifest.Files)+len(manifest.SecretFrom))-1 {
			output.WriteString("\n---\n")
		}
	}

	for i, sf := range manifest.SecretFrom {
		fileResults, err := decryptAll(&g, sf.Files, func(file string) (keyData, error) {
			key, path := fileKeyPath(file)
			data, err := decryptFile(path)
			if err != nil {
				return keyData{}, fmt.Errorf("error decrypting file %q from secretFrom.Files: %w", path, err)
			}
			return keyData{key: key, data: data}, nil
		})
		if err != nil {
			return "", err
		}

		binaryResults, err := decryptAll(&g, sf.BinaryFiles, func(file string) (keyData, error) {
			key, path := fileKeyPath(file)
			data, err := decryptFile(path)
			if err != nil {
				return keyData{}, fmt.Errorf("error decrypting file %q from secretFrom.BinaryFiles: %w", path, err)
			}
			return keyData{key: key, data: data}, nil
		})
		if err != nil {
			return "", err
		}

		envResults, err := decryptAll(&g, sf.Envs, func(file string) (keyData, error) {
			data, err := decryptFile(file)
			if err != nil {
				return keyData{}, fmt.Errorf("error decrypting file %q from secretFrom.Envs: %w", file, err)
			}
			return keyData{key: file, data: data}, nil
		})
		if err != nil {
			return "", err
		}

		stringData := make(map[string]string)
		binaryData := make(map[string]string)

		for _, r := range fileResults {
			stringData[r.key] = string(r.data)
		}
		for _, r := range binaryResults {
			binaryData[r.key] = base64.StdEncoding.EncodeToString(r.data)
		}
		for _, r := range envResults {
			env, err := godotenv.Unmarshal(string(r.data))
			if err != nil {
				return "", fmt.Errorf("error unmarshalling .env file %q: %w", r.key, err)
			}
			for k, v := range env {
				stringData[k] = v
			}
		}

		s := kubernetesSecret{
			APIVersion: "v1",
			Kind:       "Secret",
			Metadata:   sf.Metadata,
			Type:       sf.Type,
			StringData: stringData,
			Data:       binaryData,
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
	b, err := os.ReadFile(file)
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
		fmt.Fprintf(os.Stderr, "invalid format in file generator %s\n", file)
		os.Exit(1)
	}
	return slices[0], slices[1]
}
