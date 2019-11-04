// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"io/ioutil"
	"path"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

/*
Based off of
https://github.com/monopole/sopsencodedsecrets/blob/master/SopsEncodedSecrets_test.go

*/
const localTestDir = "./test/"
const pluginTestDir = "/app"

const kustomizePluginOwner = "viaduct.ai"
const kustomizePluginVersion = "v1"
const kustomizePluginName = "ksops"

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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
kustomizePluginOwner, kustomizePluginVersion, kustomizePluginName)

	th := kusttest_test.NewKustTestPluginHarness(t, pluginTestDir)

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
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
kustomizePluginOwner, kustomizePluginVersion, kustomizePluginName)

	th := kusttest_test.NewKustTestPluginHarness(t, pluginTestDir)

	// Load files from testing directory
	for _, v := range resourceVersions {
		// Write encrypt file to make it available to the test harness
		resourceName := encryptedResourceName + "-" v + encryptionSuffix +  yamlSuffix

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
