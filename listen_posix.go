// +build linux darwin

package daemon

import (
	"syscall"
)

func dup(fd int) (int, error) {
	return syscall.Dup(fd)
}
