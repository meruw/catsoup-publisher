//go:build !windows

package delivery

import "os/exec"

func hideProcess(cmd *exec.Cmd) {}
