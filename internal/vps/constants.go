package vps

import "time"

const (
	DefaultTimeout = 30 * time.Second

	shellName = "bash"
	shellFlag = "-c"

	opScriptExecution = "script execution"
	opCommand         = "command %s"

	errJsonUnmarshal = "json unmarshal: %w, response: %s"
	errTimeout       = "%s: timeout after %v"
	errCancelled     = "%s: cancelled"
	errWithStderr    = "%s: %w, stderr: %s"
	errGeneric       = "%s: %w"

	truncateSuffix = "..."
)
