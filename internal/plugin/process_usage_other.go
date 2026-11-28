//go:build !linux && !darwin

package plugin

import "os"

func processPeakRSS(_ *os.ProcessState) int64 { return 0 }
