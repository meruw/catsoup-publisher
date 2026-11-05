package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
)

// PythonValidator is a transitional adapter to the existing authoritative contract.
// It executes a fixed repository script, never a command supplied in a manifest.
type PythonValidator struct{ Root string }

func (p PythonValidator) Check(ctx context.Context, path string) (Verdict, error) {
	var result Verdict
	python := filepath.Join(p.Root, ".venv", "Scripts", "python.exe")
	script := filepath.Join(p.Root, "apps", "publisher", "support", "validate_delivery.py")
	cmd := exec.CommandContext(ctx, python, "-I", "-B", script, path)
	hideProcess(cmd)
	cmd.Dir = p.Root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return result, errors.New("delivery validation timed out")
		}
		return result, errors.New("shared validator failed; check the workspace Python environment and delivery")
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return result, errors.New("invalid response from shared validator")
	}
	return result, nil
}
