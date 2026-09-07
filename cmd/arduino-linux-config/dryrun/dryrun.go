// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// Package dryrun reports the effects a command would apply.
package dryrun

import (
	"strings"
)

type Result struct {
	// Names what the command would change, for example "carrier 'media-carrier'".
	Subject        string   `json:"subject,omitempty"`
	RebootRequired bool     `json:"reboot_required"`
	Effects        []string `json:"effects"`
}

func (r Result) Data() interface{} {
	return r
}

func (r Result) String() string {
	header := "Dry-run: no changes applied"
	if r.Subject != "" {
		header += " for " + r.Subject
	}
	rebootLine := "Reboot required: no"
	if r.RebootRequired {
		rebootLine = "Reboot required: yes"
	}
	return strings.Join(append([]string{header, rebootLine}, r.Effects...), "\n")
}
