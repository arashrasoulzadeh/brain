//go:build linux || darwin

package plugin

import (
	"os"
	"runtime"
	"syscall"
)

func processPeakRSS(state *os.ProcessState) int64 {
	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok || usage == nil || usage.Maxrss <= 0 {
		return 0
	}
	bytes := int64(usage.Maxrss)
	if runtime.GOOS == "linux" {
		bytes *= 1024
	}
	return bytes
}
