//go:build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	dockerImageName = "arduino-linux-config-test"
	containerName   = "arduino-linux-config-test-container"
)

var arch = runtime.GOARCH

func buildDockerImage(t testing.TB) {
	if t != nil {
		t.Helper()
	}
	root := repoRoot()
	cmd := exec.Command("docker", "build",
		"--platform", "linux/"+arch,
		"--build-arg", "ARCH="+arch,
		"-t", dockerImageName,
		"-f", filepath.Join(root, "tests/integration/Dockerfile"),
		root,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if t != nil {
			t.Fatalf("failed to build docker image: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to build docker image: %v\n", err)
			os.Exit(1)
		}
	}
}

func repoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get repo root: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func startDockerContainer(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "run", "-d", "--name", containerName, dockerImageName, "sleep", "infinity")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err, "failed to start docker container")
}

func stopDockerContainer(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "rm", "-f", containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func execInContainer(t *testing.T, args ...string) string {
	t.Helper()
	dockerArgs := append([]string{"exec", containerName}, args...)
	cmd := exec.Command("docker", dockerArgs...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to exec in container: %v\noutput: %s", args, string(out))
	return string(out)
}
