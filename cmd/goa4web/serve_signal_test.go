package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

const signalHelperEnv = "GOA4WEB_SIGNAL_HELPER"

func TestServeSignalHelperProcess(t *testing.T) {
	mode := os.Getenv(signalHelperEnv)
	if mode == "" {
		return
	}
	ctx, stop := notifyShutdownContext(context.Background())
	defer stop()
	fmt.Println("ready")
	<-ctx.Done()
	fmt.Println("graceful")
	if mode == "double" {
		select {}
	}
}

func TestServeFirstInterruptAllowsCleanExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	cmd, lines := startSignalHelper(t, "single")
	wantLine(t, lines, "ready")
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send first interrupt: %v", err)
	}
	wantLine(t, lines, "graceful")
	if err := waitProcess(cmd, 2*time.Second); err != nil {
		t.Fatalf("helper did not exit successfully after one interrupt: %v", err)
	}
}

func TestServeSecondInterruptForcesPromptExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	cmd, lines := startSignalHelper(t, "double")
	wantLine(t, lines, "ready")
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send first interrupt: %v", err)
	}
	wantLine(t, lines, "graceful")
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send second interrupt: %v", err)
	}
	if err := waitProcess(cmd, 2*time.Second); err == nil {
		t.Fatal("helper exited successfully; want forced signal termination")
	}
}

func startSignalHelper(t *testing.T, mode string) (*exec.Cmd, <-chan string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestServeSignalHelperProcess$")
	cmd.Env = append(os.Environ(), signalHelperEnv+"="+mode)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("create helper stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	lines := make(chan string, 2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	return cmd, lines
}

func wantLine(t *testing.T, lines <-chan string, want string) {
	t.Helper()
	select {
	case got := <-lines:
		if got != want {
			t.Fatalf("helper output = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for helper output %q", want)
	}
}

func waitProcess(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		waitErr := <-done
		return fmt.Errorf("process did not exit within %s; wait after kill: %v", timeout, waitErr)
	}
}
