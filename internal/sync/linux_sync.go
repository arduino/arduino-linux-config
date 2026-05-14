//go:build linux

// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later
package sync

import "golang.org/x/sys/unix"

func SyncToDisk() {
	unix.Sync()
}
