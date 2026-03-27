package runtime

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"example.com/test/internal/domain"
)

type Executor interface {
	Execute(command string) (string, error)
}

type WindowsExecutor struct{}

func (w *WindowsExecutor) Execute(command string) (string, error) {
	cmd := exec.Command("cmd", "/C", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

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

func ExecuteJob(job *domain.Job, executor Executor) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultCh := make(chan struct {
		output string
		err    error
	}, 1)

	go func() {
		output, err := executor.Execute(job.Command)
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
