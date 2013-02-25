// +build linux darwin

package daemon

import (
	"os"
	"syscall"
)

var signals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGHUP,
	syscall.SIGUSR1,
}

func sigAction(sig os.Signal) int {
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		return sigShutdown
	case syscall.SIGHUP:
		return sigRestart
	case syscall.SIGUSR1:
		return sigStackDump
	}
	return sigUnknown
}
