//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

const (
	dockerImageName = "arduino-linux-config-test"
	containerName   = "arduino-linux-config-test-container"
)

func buildDockerImage(t *testing.T) {
	t.Helper()
	root := FindRepositoryRootPath(t).String()
	cmd := exec.Command("docker", "build",
		"-t", dockerImageName,
		"-f", filepath.Join(root, "tests/integration/Dockerfile"),
		root,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err, "failed to build docker image")
}

func FindRepositoryRootPath(t *testing.T) *paths.Path {
	t.Helper()
	repoRootPath, err := paths.Getwd()
	require.NoError(t, err)
	for !repoRootPath.Join(".git").Exist() {
		parent := repoRootPath.Parent()
		require.NotEqual(t, parent.String(), repoRootPath.String(),
			"could not find repository root: reached filesystem root without finding .git")
		repoRootPath = parent
	}
	return repoRootPath
}

func startDockerContainer(t *testing.T) {
	t.Helper()
	buildDockerImage(t)
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
