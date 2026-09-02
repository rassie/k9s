// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

//go:build windows

package view

// initDumpSignal is a no-op on Windows, which has no SIGUSR1.
func initDumpSignal() {}
