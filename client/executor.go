// Now we will add the executor interface
package main

import (
	"os/exec"
	"runtime"
)

type Executor interface {
	Execute(command string) (string, error)
}

// Lets implement the windows executor
type WindowsExecutor struct{}

func (w *WindowsExecutor) Execute(command string) (string, error) {
	cmd := exec.Command("cmd", "/C", command)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Now the Unix one

type LinuxExecutor struct{}

func (l *LinuxExecutor) Execute(command string) (string, error) {
	cmd := exec.Command("sh", "-C", command)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func NewExecutor() Executor {
	switch runtime.GOOS {
	case "windows":
		return &WindowsExecutor{}
	default:
		return &LinuxExecutor{}
	}
}
