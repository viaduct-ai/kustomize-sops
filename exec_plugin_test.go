package main_test

import (
	"io/ioutil"
	"path"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// TestKSOPSExecPlugin uses a test file with an absolute path from the root of the repo.
// This is not necessary when using the exec plugin with kustomize
func TestKSOPSExecPlugin(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)

	defer th.Reset()

	th.ResetLoaderRoot(pluginTestDir)

	// Load files from testing directory
	generatorResource, err := ioutil.ReadFile(path.Join(localTestDir, "exec-generate-multiple-resources.yaml"))
	check(err)

	decryptedResource, err := ioutil.ReadFile(path.Join(localTestDir, decryptedMultipleResourceFile))
	check(err)

	m := th.LoadAndRunGenerator(string(generatorResource))

	th.AssertActualEqualsExpected(m, string(decryptedResource))
}
