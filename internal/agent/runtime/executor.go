package runtime

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"example.com/test/internal/domain"
)

type Executor interface {
	Execute(command string, shellType string) (string, error)
}

type WindowsExecutor struct{}

func (w *WindowsExecutor) Execute(command string, shellType string) (string, error) {
	var cmd *exec.Cmd
	switch strings.ToLower(shellType) {
	case "cmd":
		// cmd /C runs the command and exits
		cmd = exec.Command("cmd", "/C", command)

	case "powershell":
		// powershell.exe -Command runs the command
		cmd = exec.Command("powershell", "-Command", command)
	case "custom":
		cmd = exec.Command(command)

	default:
		cmd = exec.Command("cmd", "/C", command)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type LinuxExecutor struct{}

func (l *LinuxExecutor) Execute(command string, shellType string) (string, error) {
	var cmd *exec.Cmd
	switch strings.ToLower(shellType) {
	case "sh":
		// sh -c runs the command and exits
		cmd = exec.Command("sh", "-c", command)

	case "bash":
		// bash -c runs the command and exits
		cmd = exec.Command("bash", "-c", command)
	case "custom":
		cmd = exec.Command(command)

	default:
		cmd = exec.Command("sh", "-c", command)
	}
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

func ExecuteJob(job *domain.Job, executor Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	resultCh := make(chan struct {
		output string
		err    error
	}, 1)

	go func() {
		output, err := executor.Execute(job.Command, job.ShellType)
		resultCh <- struct {
			output string
			err    error
		}{output: output, err: err}
	}()

	select {
	case <-ctx.Done():
		job.Output = ctx.Err().Error()
		job.Status = domain.FAILED
	case result := <-resultCh:
		job.Output = result.output
		if result.err != nil {
			job.Status = domain.FAILED
			return
		}
		job.Status = domain.FINISHED
	}
}
