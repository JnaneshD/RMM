package api

type PushMessageRequest struct {
	Command   string `json:"command" binding:"required"`
	ShellType string `json:"shell_type" binding:"required"`
	JobDir    string `json:"job_dir"`
}

type RegisterRequest struct {
	UUID        string `json:"uuid" binding:"required"`
	FingerPrint string `json:"fingerprint" binding:"required"`
	TimeStamp   string `json:"timestamp"`
	Signature   string `json:"signature" binding:"required"`
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
}
