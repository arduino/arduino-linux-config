// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build linux

package sync

import "syscall"

func SyncToDisk() {
	syscall.Sync()
}
