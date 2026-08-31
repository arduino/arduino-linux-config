// This file is part of arduino-linux-config.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package executor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-linux-config/internal/sync"
)

// Executor performs every side effect of a command.
// Both the real and the dry-run implementation walk the same code path.
type Executor interface {
	Run(ctx context.Context, args ...string) error
	MkdirAll(dir *paths.Path) error
	Rename(from, to *paths.Path) error
	Remove(file *paths.Path) error
	WriteFile(file *paths.Path, data []byte, perm os.FileMode) error
	Sync()
}

func Real() Executor {
	return realExecutor{}
}

// NewRecorder returns a dry-run Executor: it applies nothing and collects the effects.
func NewRecorder() *Recorder {
	return &Recorder{}
}

type realExecutor struct{}

func (realExecutor) Run(ctx context.Context, args ...string) error {
	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return fmt.Errorf("failed to create process %v: %w", args, err)
	}

	if _, stderr, err := cmd.RunAndCaptureOutput(ctx); err != nil {
		return fmt.Errorf("command %v failed: %w (stderr: %s)", args, err, stderr)
	}
	return nil
}

func (realExecutor) MkdirAll(dir *paths.Path) error {
	return dir.MkdirAll()
}

func (realExecutor) Rename(from, to *paths.Path) error {
	return from.Rename(to)
}

func (realExecutor) Remove(file *paths.Path) error {
	return file.Remove()
}

func (realExecutor) WriteFile(file *paths.Path, data []byte, perm os.FileMode) error {
	return os.WriteFile(file.String(), data, perm)
}

func (realExecutor) Sync() {
	sync.SyncToDisk()
}

type Recorder struct {
	effects []string
}

func (r *Recorder) Effects() []string {
	return r.effects
}

func (r *Recorder) Run(_ context.Context, args ...string) error {
	r.record(quote(args...))
	return nil
}

func (r *Recorder) MkdirAll(dir *paths.Path) error {
	r.record(quote("mkdir", "-p", dir.String()))
	return nil
}

func (r *Recorder) Rename(from, to *paths.Path) error {
	r.record(quote("mv", from.String(), to.String()))
	return nil
}

func (r *Recorder) Remove(file *paths.Path) error {
	r.record(quote("rm", "-f", file.String()))
	return nil
}

func (r *Recorder) WriteFile(file *paths.Path, data []byte, _ os.FileMode) error {
	r.record(fmt.Sprintf("# write %s (%d bytes)", quote(file.String()), len(data)))
	return nil
}

func (r *Recorder) Sync() {
	r.record("sync")
}

func (r *Recorder) record(effect string) {
	r.effects = append(r.effects, effect)
}

// Quotes only the arguments that need it.
func quote(args ...string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = arg
		if arg == "" || strings.ContainsAny(arg, " \t\n'\"\\$`") {
			quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
		}
	}
	return strings.Join(quoted, " ")
}
