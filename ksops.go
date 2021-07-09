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

	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/decrypt"
	"sigs.k8s.io/yaml"
)

type ksops struct {
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
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

	if manifest.Files == nil {
		fmt.Fprintln(os.Stderr, "missing the required 'files' key in the ksops manifests")
		os.Exit(1)
	}

	var output bytes.Buffer

	for _, file := range manifest.Files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %q: %q\n", file, err.Error())
			os.Exit(1)
		}

		format := formats.FormatForPath(file)
		data, err := decrypt.DataWithFormat(b, format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "trouble decrypting file %s", err.Error())
			os.Exit(1)
		}

		output.Write(data)
		output.WriteString("\n---\n")
	}

	fmt.Print(output.String())
}
