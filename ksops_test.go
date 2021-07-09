// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"io/ioutil"
	"path"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

const localTestDir = "./test/"
const pluginTestDir = "/app"

const yamlSuffix = ".yaml"
const encryptionSuffix = ".enc"

const generatorSingleResourceFile = "generate-single-resource.yaml"

const encryptedResourceName = "secret"
const encryptedResourceFile = encryptedResourceName + encryptionSuffix + yamlSuffix

const decryptedSingleResourceFile = "secret.yaml"

const generatorMultipleResourcesFile = "generate-multiple-resources.yaml"
const decryptedMultipleResourceFile = "multiple-secrets.yaml"

var resourceVersions = [3]string{"A", "B", "C"}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestKSOPSPluginSingleResource(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)

	defer th.Reset()

	th.ResetLoaderRoot(pluginTestDir)

	// Load files from testing directory
	encryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, encryptedResourceFile))
	check(err)

	generatorResource, err := ioutil.ReadFile(path.Join(localTestDir, generatorSingleResourceFile))
	check(err)

	decryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, decryptedSingleResourceFile))
	check(err)

	// Write encrypt file to make it available to the test harness
	th.WriteF(path.Join(pluginTestDir, encryptedResourceFile), string(encryptedResource))

	m := th.LoadAndRunGenerator(string(generatorResource))

	th.AssertActualEqualsExpected(m, string(decryptedResource))
}

func TestKSOPSPluginMultipleResources(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.ResetLoaderRoot(pluginTestDir)

	// Load files from testing directory
	for _, v := range resourceVersions {
		// Write encrypt file to make it available to the test harness
		resourceName := encryptedResourceName + "-" + v + encryptionSuffix + yamlSuffix

		encryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, resourceName))
		check(err)

		th.WriteF(path.Join(pluginTestDir, resourceName), string(encryptedResource))

	}
	generatorResource, err := ioutil.ReadFile(path.Join(localTestDir, generatorMultipleResourcesFile))
	check(err)

	decryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, decryptedMultipleResourceFile))
	check(err)

	m := th.LoadAndRunGenerator(string(generatorResource))

	th.AssertActualEqualsExpected(m, string(decryptedResource))
}

func TestKSOPSPluginHashAnnotation(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)

	defer th.Reset()

	th.ResetLoaderRoot(pluginTestDir)

	// Load files from testing directory
	encryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, "hash", encryptedResourceFile))
	check(err)

	generatorResource, err := ioutil.ReadFile(path.Join(localTestDir, "hash", "generate-resources.yaml"))
	check(err)

	// Write encrypt file to make it available to the test harness
	th.WriteF(path.Join(pluginTestDir, encryptedResourceFile), string(encryptedResource))

	m := th.LoadAndRunGenerator(string(generatorResource))

	if m.Resources()[0].NeedHashSuffix() != true {
		t.Errorf("expected resource to need hashing")
	}
}

func TestKSOPSPluginBehaviorAnnotation(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)

	defer th.Reset()

	th.ResetLoaderRoot(pluginTestDir)

	// Load files from testing directory
	encryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, "behaviors", encryptedResourceFile))
	check(err)

	generatorResource, err := ioutil.ReadFile(path.Join(localTestDir, "behaviors", "generate-resources.yaml"))
	check(err)

	// Write encrypt file to make it available to the test harness
	th.WriteF(path.Join(pluginTestDir, encryptedResourceFile), string(encryptedResource))

	m := th.LoadAndRunGenerator(string(generatorResource))

	resource := m.Resources()[0]

	if resource.Behavior() != types.BehaviorReplace {
		t.Errorf("expected resource to have behavior %d (replace), but has %d", types.BehaviorReplace, resource.Behavior())
	}
}
