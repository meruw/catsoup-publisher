//go:build !windows

package mediahost

import "os/exec"

func hideProcess(cmd *exec.Cmd) {}
