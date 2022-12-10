/*
Copyright 2019 viaduct.ai
SPDX-License-Identifier: Apache-2.0

KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

KSOPS, or kustomize-SOPS, is a kustomize plugin for SOPS encrypted resources. KSOPS can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a kustomize plugin, KSOPS allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.
*/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/decrypt"
	"sigs.k8s.io/yaml"
)

type metadata struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type secret struct {
	ApiVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   metadata          `json:"metadata" yaml:"metadata"`
	Type       string            `json:"type,omitempty" yaml:"type,omitempty"`
	StringData map[string]string `json:"stringData" yaml:"stringData"`
}

type secretFrom struct {
	Files    []string `json:"files,omitempty" yaml:"files,omitempty"`
	Envs     []string `json:"envs,omitempty" yaml:"envs,omitempty"`
	Metadata metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
}

type ksops struct {
	Files      []string     `json:"files,omitempty" yaml:"files,omitempty"`
	SecretFrom []secretFrom `json:"secretFrom,omitempty" yaml:"secretFrom,omitempty"`
}

func decryptFile(file string, content []byte) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %q: %q\n", file, err.Error())
		fmt.Fprintf(os.Stderr, "manifest content: %s", content)
		os.Exit(1)
	}

	format := formats.FormatForPath(file)
	data, err := decrypt.DataWithFormat(b, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trouble decrypting file %s", err.Error())
		os.Exit(1)
	}
	return data
}

// main executes KOSPS as an exec plugin
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "received too few args:", os.Args)
		fmt.Fprintln(os.Stderr, "always invoke this via kustomize plugins")
		os.Exit(1)
	}

	// ignore the first file name argument
	// load the second argument, the file path
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read in manifest", os.Args[1])
		os.Exit(1)
	}

	var manifest ksops
	err = yaml.Unmarshal(content, &manifest)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error unmarshalling manifest content: %q \n%s\n", err, content)
		os.Exit(1)
	}

	if manifest.Files == nil && manifest.SecretFrom == nil {
		fmt.Fprintf(os.Stderr, "missing the required 'files' or 'secretFrom' key in the ksops manifests: %s", content)
		os.Exit(1)
	}

	var output bytes.Buffer

	for _, file := range manifest.Files {
		data := decryptFile(file, content)

		output.Write(data)
		output.WriteString("\n---\n")
	}

	for _, secretFrom := range manifest.SecretFrom {
		stringData := make(map[string]string)

		for _, file := range secretFrom.Files {
			data := decryptFile(file, content)

			key := filepath.Base(file)
			stringData[key] = string(data)
		}

		for _, file := range secretFrom.Envs {
			data := decryptFile(file, content)

			env, err := godotenv.Unmarshal(string(data))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error unmarshalling .env file %s", err.Error())
				os.Exit(1)
			}
			for k, v := range env {
				stringData[k] = v
			}
		}

		s := secret{
			ApiVersion: "v1",
			Kind:       "Secret",
			Metadata:   secretFrom.Metadata,
			Type:       secretFrom.Type,
			StringData: stringData,
		}
		d, err := yaml.Marshal(&s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "trouble encoding metadata %s", err.Error())
		}
		output.WriteString(string(d))
		output.WriteString("---\n")
	}

	fmt.Print(output.String())
}
