// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"

	"github.com/pkg/errors"
	"go.mozilla.org/sops/decrypt"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/yaml"
)

// Loads and decrypts sops-encoded files
// References https://github.com/mozilla/sops
// Based on https://github.com/monopole/sopsencodedsecrets
type plugin struct {
	rf    *resmap.Factory
	ldr   ifc.Loader
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`
}

//noinspection GoUnusedGlobalVariable
//nolint: golint
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	p.rf = rf
	p.ldr = ldr
	return yaml.Unmarshal(c, p)
}

func (p *plugin) loadAndDecryptFile(f string) ([]byte, error) {
	bytes, err := p.ldr.Load(f)

	if err != nil {
		return bytes, errors.Wrapf(err, "trouble reading file %s", f)
	}

	return decrypt.Data(bytes, "yaml")
}

func (p *plugin) loadDecryptedResources() (resmap.ResMap, error) {
	resourcesBuffer := bytes.Buffer{}
	numResources := len(p.Files)

	for i, f := range p.Files {

		decryptedBytes, err := p.loadAndDecryptFile(f)

		if err != nil {
			return nil, errors.Wrapf(err, "trouble decrypting file %s", f)
		}

		resourcesBuffer.Write(decryptedBytes)

		// Do not add document end for a single resource or the last resource
		if numResources > 1 && i != numResources-1 {
			// Check if this is valid at the top of a resource (start with ---)
			resourcesBuffer.WriteString("---\n")
		}
	}

	resourcesBytes := resourcesBuffer.Bytes()

	resMap, err := p.rf.NewResMapFromBytes(resourcesBytes)

	if err != nil {
		return nil, errors.Wrapf(err, "trouble converting bytes to ResMap %s", resourcesBuffer.String())
	}

	return resMap, err
}

// Generate the output. A KustomizePlugin will call this generate function
func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.loadDecryptedResources()
}
