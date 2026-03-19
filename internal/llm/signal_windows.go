//go:build windows

package llm

import "os"

func terminateProcess(process *os.Process) error {
	return process.Kill()
}
