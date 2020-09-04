/*
Copyright 2019 viaduct.ai
SPDX-License-Identifier: Apache-2.0

KSOPS - A Flexible Kustomize Plugin for SOPS Encrypted Resource

KSOPS, or kustomize-SOPS, is a kustomize plugin for SOPS encrypted resources. KSOPS can be used to decrypt any Kubernetes resource, but is most commonly used to decrypt encrypted Kubernetes Secrets and ConfigMaps. As a kustomize plugin, KSOPS allows you to manage, build, and apply encrypted manifests the same way you manage the rest of your Kubernetes manifests.
*/
package main

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"go.mozilla.org/sops/v3/decrypt"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	hashAnnotation     = "kustomize.config.k8s.io/needs-hash"
	behaviorAnnotation = "kustomize.config.k8s.io/behavior"
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

// Generate the output. A KustomizePlugin will call this generate function
func (p *plugin) Generate() (resmap.ResMap, error) {

	// get a decrypted resmap for each files
	var resources resmap.ResMap

	// validate the files key exists
	if p.Files == nil {
		return nil, errors.New("missing the required 'files' key in the ksops manifests")
	}

	for _, f := range p.Files {
		// check for err
		r, err := decryptResource(p, f)

		// fail hard on any error
		if err != nil {
			return nil, errors.Wrapf(err, "trouble converting file '%s' to a resource", f)
		}

		// absorb all (i.e merge/replace colliding IDs)
		if resources != nil {
			err = resources.AbsorbAll(r)
			if err != nil {
				return nil, errors.Wrapf(err, "trouble absorbing resources described in '%s'", f)
			}
		} else {
			// initialize first resmap
			resources = r
		}
	}

	//https://github.com/kubernetes-sigs/kustomize/blob/master/plugin/builtin/hashtransformer/HashTransformer.go
	return UpdateResourceOptions(resources)
}

// UpdateResourceOptions updates (mutates) the resource's options based off the kustomize annotations
// Taken from the kustomize ExecPlugin struct
// https://github.com/kubernetes-sigs/kustomize/blob/4fc859b62eb5f61e5a01ec1d8d365c87f4771fd5/api/internal/plugins/execplugin/execplugin.go#L246
func UpdateResourceOptions(rm resmap.ResMap) (resmap.ResMap, error) {
	for _, r := range rm.Resources() {
		// Disable name hashing by default and require plugin to explicitly
		// request it for each resource.
		annotations := r.GetAnnotations()
		behavior := annotations[behaviorAnnotation]
		var needsHash bool
		if val, ok := annotations[hashAnnotation]; ok {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf(
					"the annotation %q contains an invalid value (%q)",
					hashAnnotation, val)
			}
			needsHash = b
		}
		delete(annotations, hashAnnotation)
		delete(annotations, behaviorAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)
		r.SetOptions(types.NewGenArgs(
			&types.GeneratorArgs{
				Behavior: behavior,
				Options: &types.GeneratorOptions{
					DisableNameSuffixHash: !needsHash,
				},
			}))
	}
	return rm, nil
}

func decryptResource(p *plugin, f string) (resmap.ResMap, error) {
	b, err := decryptFile(p, f)

	if err != nil {
		return nil, errors.Wrapf(err, "trouble decrypting file %s", f)
	}

	return p.rf.NewResMapFromBytes(b)
}

func decryptFile(p *plugin, f string) ([]byte, error) {
	b, err := p.ldr.Load(f)

	if err != nil {
		return b, errors.Wrapf(err, "trouble reading file %s", f)
	}

	return decrypt.Data(b, "yaml")
}

func main() {}
