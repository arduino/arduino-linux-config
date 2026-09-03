// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

const (
	dockerImageName = "arduino-linux-config-test"
	containerName   = "arduino-linux-config-test-container"

	ventunoqDockerImageName = "arduino-linux-config-test-ventunoq"
	ventunoqContainerName   = "arduino-linux-config-test-container-ventunoq"

	// The device tree partition, emulated with a loop-mounted FAT image.
	ventunoqDtbImage     = "/dtb_a.img"
	ventunoqDtbPartition = "/dev/disk/by-partlabel/dtb_a"
)

func buildDockerImage(t *testing.T, imageName, dockerfile string) {
	t.Helper()
	root := FindRepositoryRootPath(t).String()
	cmd := exec.Command("docker", "build",
		"-t", imageName,
		"-f", filepath.Join(root, dockerfile),
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

func runDockerContainer(t *testing.T, imageName, name string, runArgs ...string) {
	t.Helper()
	args := append([]string{"run", "-d", "--name", name}, runArgs...)
	args = append(args, imageName, "sleep", "infinity")
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err, "failed to start docker container")
}

func startDockerContainer(t *testing.T) {
	t.Helper()
	buildDockerImage(t, dockerImageName, "tests/integration/Dockerfile")
	runDockerContainer(t, dockerImageName, containerName)
}

func stopDockerContainer(t *testing.T) {
	t.Helper()
	removeDockerContainer(t, containerName)
}

// VentunoQ writes the device tree on a vfat partition, so the container needs the
// privileges to set up a loop device and to mount it.
func startVentunoqDockerContainer(t *testing.T) {
	t.Helper()
	buildDockerImage(t, ventunoqDockerImageName, "tests/integration/ventunoq.Dockerfile")
	runDockerContainer(t, ventunoqDockerImageName, ventunoqContainerName, "--privileged")

	execInVentunoqContainer(t, "sh", "-c", strings.Join([]string{
		"set -eu",
		"dd if=/dev/zero of=" + ventunoqDtbImage + " bs=1M count=16",
		"mkfs.vfat " + ventunoqDtbImage,
		"mkdir -p " + filepath.Dir(ventunoqDtbPartition),
		"loop=$(losetup --find --show " + ventunoqDtbImage + ")",
		`ln -sf "$loop" ` + ventunoqDtbPartition,
	}, "\n"))
}

// The image fakes an Ubuntu root, the only distribution supported on VentunoQ.
func startVentunoqUbuntuDockerContainer(t *testing.T) {
	t.Helper()
	buildDockerImage(t, ventunoqDockerImageName, "tests/integration/ventunoq.Dockerfile")
	runDockerContainer(t, ventunoqDockerImageName, ventunoqContainerName)
}

func startVentunoqDebianDockerContainer(t *testing.T) {
	t.Helper()
	startVentunoqUbuntuDockerContainer(t)
	execInVentunoqContainer(t, "sh", "-c", `printf 'ID=debian\n' > /tmp/compat-root/etc/os-release`)
}

func stopVentunoqDockerContainer(t *testing.T) {
	t.Helper()
	// The loop device lives in the host kernel, detach it before dropping the container.
	_ = exec.Command("docker", "exec", ventunoqContainerName, "sh", "-c",
		"losetup -j "+ventunoqDtbImage+" | cut -d: -f1 | xargs -r losetup -d").Run()
	removeDockerContainer(t, ventunoqContainerName)
}

func removeDockerContainer(t *testing.T, name string) {
	t.Helper()
	cmd := exec.Command("docker", "rm", "-f", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func execInContainer(t *testing.T, args ...string) string {
	t.Helper()
	return execInNamedContainer(t, containerName, args...)
}

func execInVentunoqContainer(t *testing.T, args ...string) string {
	t.Helper()
	return execInNamedContainer(t, ventunoqContainerName, args...)
}

func execInNamedContainer(t *testing.T, name string, args ...string) string {
	t.Helper()
	out, err := execInNamedContainerWithError(t, name, args...)
	require.NoError(t, err, "failed to exec in container: %v\noutput: %s", args, out)
	return out
}

func execInNamedContainerWithError(t *testing.T, name string, args ...string) (string, error) {
	t.Helper()
	dockerArgs := append([]string{"exec", name}, args...)
	cmd := exec.Command("docker", dockerArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
