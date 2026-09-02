// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

//go:build !windows

package view

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
)

// initDumpSignal installs a SIGUSR1 handler that writes the stacks of all
// goroutines to a file next to the k9s log. It is a diagnostic aid for UI
// hangs: run `kill -USR1 <pid>` while the app is frozen and inspect
// the dump; the app keeps running.
func initDumpSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR1)

	go func() {
		for range sig {
			dumpGoroutines()
		}
	}()
}

func dumpGoroutines() {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	dir := filepath.Dir(config.AppLogFile)
	if config.AppLogFile == "" {
		dir = os.TempDir()
	}
	path := filepath.Join(dir, fmt.Sprintf("k9s-goroutines-%s.txt", time.Now().Format("20060102-150405")))
	if err := os.WriteFile(path, buf, 0o600); err != nil {
		slog.Error("Goroutine dump failed", slogs.Path, path, slogs.Error, err)
		return
	}
	slog.Info("Goroutine dump written", slogs.Path, path)
}
