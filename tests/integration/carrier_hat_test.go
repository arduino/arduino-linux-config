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

// TestCarrierHatDryRunCommands verifies the fdtoverlay command that would be
// executed when a hat is enabled or disabled on top of a configured carrier:
// the carrier overlay must be preserved in both cases. Every assertion is done
// by re-running the previously applied command with --dry-run.
func TestCarrierHatDryRunCommands(t *testing.T) {
	startVentunoqDockerContainer(t)
	t.Cleanup(func() { stopVentunoqDockerContainer(t) })

	// Fresh install: no persisted state is expected.
	statusDir := execInVentunoqContainer(t, "ls", "-A", "/var/lib/arduino-linux-config/status")
	require.Empty(t, strings.TrimSpace(statusDir), "status directory should be empty on a fresh install")

	const (
		overlaysDir       = "/var/lib/arduino-linux-config/overlays/"
		carrierOverlay    = overlaysDir + "monaco-monza-dsi-waveshare,8.0-dsi-touch-a.dtbo"
		automationOverlay = overlaysDir + "monaco-monza-automation-hat.dtbo"
	)

	// Persist the carrier configuration; every subsequent dry-run runs on top of it.
	execInVentunoqContainer(t, "arduino-linux-config", "hw", "enable", "media-carrier", "display=8-dsi-touch-a")

	// Re-running the same command as dry-run must produce the carrier overlay only.
	result := dryRunHwCommand(t, "enable", "media-carrier", "display=8-dsi-touch-a")
	require.ElementsMatch(t, []string{carrierOverlay}, extractFdtoverlayOverlays(t, result.Effects))

	// Assert 1: enabling automation must keep the carrier overlay in the reload.
	execInVentunoqContainer(t, "arduino-linux-config", "hw", "enable", "automation")
	result = dryRunHwCommand(t, "enable", "automation")
	require.ElementsMatch(t,
		[]string{automationOverlay, carrierOverlay},
		extractFdtoverlayOverlays(t, result.Effects),
	)

	// Assert 2: disabling automation must keep the carrier overlay in the reload.
	execInVentunoqContainer(t, "arduino-linux-config", "hw", "disable", "automation")
	result = dryRunHwCommand(t, "disable", "automation")
	require.ElementsMatch(t, []string{carrierOverlay}, extractFdtoverlayOverlays(t, result.Effects))
}
