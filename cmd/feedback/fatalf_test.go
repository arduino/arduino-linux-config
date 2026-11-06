// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package feedback

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFatalf(t *testing.T) {
	const formatEnv = "ARDUINO_LINUX_CONFIG_FATALF_TEST_FORMAT"

	// Run Fatalf in a subprocess because it terminates the process.
	if outputFormat, helperProcess := os.LookupEnv(formatEnv); helperProcess {
		reset()
		selectedFormat, valid := ParseOutputFormat(outputFormat)
		require.True(t, valid)
		SetFormat(selectedFormat)

		Fatalf(ErrBadArgument, "invalid %s: %d%%", "value", 42)
		t.Fatal("Fatalf returned without exiting")
	}

	cases := []struct {
		name string
		want string
	}{
		{
			name: "text",
			want: "invalid value: 42%\n",
		},
		{
			name: "json",
			want: "{\n  \"error\": \"invalid value: 42%\"\n}\n",
		},
		{
			name: "jsonmini",
			want: "{\"error\":\"invalid value: 42%\"}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			executable, err := os.Executable()
			require.NoError(t, err)

			// #nosec G204 -- Re-execute our own test binary to check os.Exit.
			cmd := exec.Command(executable, "-test.run=^TestFatalf$")
			cmd.Env = append(os.Environ(), formatEnv+"="+tc.name)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()

			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr)
			require.Equal(t, int(ErrBadArgument), exitErr.ExitCode())
			require.Empty(t, stdout.String())
			require.Equal(t, tc.want, stderr.String())
		})
	}
}
