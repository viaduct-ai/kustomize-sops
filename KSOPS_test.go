// Copyright 2019 viaduct.ai
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"io/ioutil"
	"path"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
)

/*
Based off of
https://github.com/monopole/sopsencodedsecrets/blob/master/SopsEncodedSecrets_test.go

*/
const localTestDir = "./test/"
const pluginTestDir = "/app"

const generatorSingleResourceFile = "generateSingleResource.yaml"
const encryptedResourceFile = "encryptedSecret.yaml"
const decryptedSingleResourceFile = "secret.yaml"

const generatorMultipleResourcesFile = "generateMultipleResources.yaml"
const decryptedMultipleResourceFile = "multipleSecrets.yaml"

var resourceVersions = [3]string{"A", "B", "C"}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestKSOPSPluginSingleResource(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"viaduct.ai", "v1", "KSOPS")

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
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"viaduct.ai", "v1", "KSOPS")

	th := kusttest_test.NewKustTestPluginHarness(t, pluginTestDir)

	// Load files from testing directory
	for _, v := range resourceVersions {
		// Write encrypt file to make it available to the test harness
		resourceName := "encryptedSecret" + v + ".yaml"

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
