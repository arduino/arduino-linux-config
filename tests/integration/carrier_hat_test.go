// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build integration

package integration

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Mirrors the JSON output of a dry-run enable/disable command.
type dryRunResult struct {
	Subject string   `json:"subject"`
	Effects []string `json:"effects"`
}

// The temporary dtb name contains a nanosecond timestamp, so the fdtoverlay
// command is matched by its stable prefix.
var fdtoverlayRe = regexp.MustCompile(
	`^fdtoverlay -i /run/arduino-linux-config/dtb/monza\.dtb ` +
		`-o /run/arduino-linux-config/dtb/temporaryDeviceTree\.\d+\.temp (.*)$`,
)

func extractFdtoverlayOverlays(t *testing.T, effects []string) []string {
	t.Helper()
	var matches []string
	for _, e := range effects {
		if m := fdtoverlayRe.FindStringSubmatch(e); m != nil {
			matches = append(matches, strings.Fields(m[1])...)
		}
	}
	require.NotEmpty(t, matches, "no fdtoverlay effect found in: %v", effects)
	return matches
}

func dryRunHwCommand(t *testing.T, args ...string) dryRunResult {
	t.Helper()
	full := append([]string{"arduino-linux-config", "hw"}, args...)
	full = append(full, "--dry-run", "--format", "json")
	out := execInVentunoqContainer(t, full...)

	var result dryRunResult
	require.NoError(t, json.Unmarshal([]byte(out), &result), "output should be valid JSON: %s", out)
	return result
}
