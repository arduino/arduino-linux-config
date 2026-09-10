// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestParseBuildID(t *testing.T) {
	tests := []struct {
		in      string
		want    BuildInfo
		wantErr bool
	}{
		{in: "20260825-260", want: BuildInfo{Date: 20260825, Increment: 260}},
		{in: "20260729-248", want: BuildInfo{Date: 20260729, Increment: 248}},
		{in: "\"20260825-260\"", want: BuildInfo{Date: 20260825, Increment: 260}},
		{in: "20260825", wantErr: true},
		{in: "abc-123", wantErr: true},
		{in: "20260825-x", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseBuildID(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBuildInfoLessThan(t *testing.T) {
	min := BuildInfo{Date: 20260825, Increment: 260}
	require.True(t, BuildInfo{Date: 20260729, Increment: 248}.LessThan(min))
	require.True(t, BuildInfo{Date: 20260825, Increment: 259}.LessThan(min))
	require.False(t, BuildInfo{Date: 20260825, Increment: 260}.LessThan(min))
	require.False(t, BuildInfo{Date: 20260825, Increment: 261}.LessThan(min))
	require.False(t, BuildInfo{Date: 20260826, Increment: 1}.LessThan(min))
}

func TestGetBuildInfoFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"etc/buildinfo": &fstest.MapFile{Data: []byte("BUILD_ID=20260825-260\nOTHER=foo\n")},
	}
	got, err := getBuildInfoFromFS(fsys)
	require.NoError(t, err)
	require.Equal(t, BuildInfo{Date: 20260825, Increment: 260}, got)
}

func TestGetBuildInfoFromFS_Missing(t *testing.T) {
	fsys := fstest.MapFS{}
	_, err := getBuildInfoFromFS(fsys)
	require.Error(t, err)
}

func TestGetBuildInfoFromFS_NoBuildID(t *testing.T) {
	fsys := fstest.MapFS{
		"etc/buildinfo": &fstest.MapFile{Data: []byte("OTHER=foo\n")},
	}
	_, err := getBuildInfoFromFS(fsys)
	require.Error(t, err)
}
