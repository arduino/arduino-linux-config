// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// BuildInfo represents a parsed /etc/buildinfo BUILD_ID value.
// BUILD_ID has the form YYYYMMDD-INCREMENTAL, e.g. 20260825-260.
type BuildInfo struct {
	Date      int
	Increment int
}

// String returns the canonical BUILD_ID representation.
func (b BuildInfo) String() string {
	return fmt.Sprintf("%d-%d", b.Date, b.Increment)
}

// LessThan reports whether b is older than other.
func (b BuildInfo) LessThan(other BuildInfo) bool {
	if b.Date != other.Date {
		return b.Date < other.Date
	}
	return b.Increment < other.Increment
}

// parseBuildID parses a raw BUILD_ID string like "20260825-260".
func parseBuildID(id string) (BuildInfo, error) {
	id = strings.Trim(id, "\"' \t\n\r")
	date, incr, ok := strings.Cut(id, "-")
	if !ok {
		return BuildInfo{}, fmt.Errorf("malformed BUILD_ID %q", id)
	}
	d, err := strconv.Atoi(date)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("malformed BUILD_ID date %q: %w", date, err)
	}
	i, err := strconv.Atoi(incr)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("malformed BUILD_ID increment %q: %w", incr, err)
	}
	return BuildInfo{Date: d, Increment: i}, nil
}

// getBuildInfoFromFS reads etc/buildinfo from the given filesystem and returns
// the parsed BUILD_ID value.
func getBuildInfoFromFS(fsys fs.FS) (BuildInfo, error) {
	f, err := fsys.Open("etc/buildinfo")
	if err != nil {
		return BuildInfo{}, fmt.Errorf("cannot read /etc/buildinfo: %w", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if v, ok := strings.CutPrefix(line, "BUILD_ID="); ok {
			return parseBuildID(v)
		}
	}
	if err := s.Err(); err != nil {
		return BuildInfo{}, fmt.Errorf("cannot read /etc/buildinfo: %w", err)
	}
	return BuildInfo{}, fmt.Errorf("BUILD_ID not found in /etc/buildinfo")
}

// GetBuildInfo reads /etc/buildinfo from the root filesystem, or from
// COMPATIBLE_ROOT_DIR if set (used in integration tests).
func GetBuildInfo() (BuildInfo, error) {
	root := "/"
	if dir := os.Getenv("COMPATIBLE_ROOT_DIR"); dir != "" {
		root = dir
	}
	return getBuildInfoFromFS(os.DirFS(root))
}
