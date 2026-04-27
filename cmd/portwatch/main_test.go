package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain_Version(t *testing.T) {
	if os.Getenv("RUN_MAIN") == "1" {
		os.Args = []string{"portwatch", "-version"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_Version")
	cmd.Env = append(os.Environ(), "RUN_MAIN=1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := string(out)
	if got == "" {
		t.Error("expected version output, got empty string")
	}
}

func TestMain_MissingConfig(t *testing.T) {
	if os.Getenv("RUN_MAIN") == "1" {
		os.Args = []string{"portwatch", "-config", "/nonexistent/path/config.yaml"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MissingConfig")
	cmd.Env = append(os.Environ(), "RUN_MAIN=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing config, got nil")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}
	}
}
