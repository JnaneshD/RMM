// Lets create a job type with necessary fields
package main

import (
	"os/exec"

	"example.com/test/models"
)

func Execute(job *models.Job) {
	// Now we need to execute the job in the client
	cmd := exec.Command("cmd", "/C", job.Command)
	stdout, err := cmd.Output()

	if err != nil {
		job.Output = err.Error()
		job.Status = models.FAILED
	} else {
		job.Status = models.FINISHED
		job.Output = string(stdout)
	}
}
