//go:build !windows

package llm

import (
	"os"
	"syscall"
)

func terminateProcess(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
