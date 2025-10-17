//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildHelloWorldServlet tests that the engine-java binary can successfully
// build a simple Hello World Tomcat servlet project
func TestBuildHelloWorldServlet(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Determine binary name based on OS and architecture
	binaryName := fmt.Sprintf("engine-java-%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryPath := filepath.Join(cwd, binaryName)

	// Check if pre-built binary exists, if not build it
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Logf("Binary %s not found, building it...", binaryName)
		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		err = cmd.Run()
		require.NoError(t, err, "Failed to build binary")
	}

	// Path to test project
	testProjectPath := filepath.Join(cwd, "testdata", "hello-world-servlet")

	// Ensure test project exists
	pomPath := filepath.Join(testProjectPath, "pom.xml")
	require.FileExists(t, pomPath, "Test project pom.xml not found")

	// Clean any previous build artifacts
	targetDir := filepath.Join(testProjectPath, "target")
	os.RemoveAll(targetDir)
	defer os.RemoveAll(targetDir)

	// Change to project directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(testProjectPath)
	require.NoError(t, err)

	// Run the binary to build the test project
	t.Log("Running engine-java to build Hello World servlet...")
	cmd := exec.Command(binaryPath, "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CONTAINIFYCI_FILE=.containifyci/containifyci.go")

	// Run command and check exit code
	err = cmd.Run()
	assert.NoError(t, err, "Build command should exit with code 0")

	t.Log("Build completed successfully with exit code 0")
}
