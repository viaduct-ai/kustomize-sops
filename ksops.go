/*
Copyright 2019 viaduct.ai
SPDX-License-Identifier: Apache-2.0

KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

KSOPS, or kustomize-SOPS, is a kustomize plugin for SOPS encrypted resources. KSOPS can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a kustomize plugin, KSOPS allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.
*/
package main

import (
	"bytes"

	"github.com/pkg/errors"
	"go.mozilla.org/sops/v3/decrypt"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
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

// KustomizePlugin export to satisfy the kustomize Go plugin interface
//noinspection GoUnusedGlobalVariable
//nolint: golint
var KustomizePlugin plugin

func (p *plugin) Config(
	ph *resmap.PluginHelpers, c []byte) error {
	p.rf = ph.ResmapFactory()
	p.ldr = ph.Loader()
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

	for _, f := range p.Files {

		decryptedBytes, err := p.loadAndDecryptFile(f)

		if err != nil {
			return nil, errors.Wrapf(err, "trouble decrypting file %s", f)
		}

		// Write and separate resources
		resourcesBuffer.Write(decryptedBytes)
		resourcesBuffer.WriteString("---\n")
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

func main() {}
