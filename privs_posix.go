// +build linux darwin

package daemon

import (
	"os/user"
	"strconv"
	"syscall"
)

func chuser(username string) (uid, gid int) {
	usr, err := user.Lookup(username)
	if err != nil {
		Fatal.Printf("failed to find user %q: %s", username, err)
	}

	uid, err = strconv.Atoi(usr.Uid)
	if err != nil {
		Fatal.Printf("bad user ID %q: %s", usr.Uid, err)
	}

	gid, err = strconv.Atoi(usr.Gid)
	if err != nil {
		Fatal.Printf("bad group ID %q: %s", usr.Gid, err)
	}

	if err := syscall.Setgid(gid); err != nil {
		Fatal.Printf("setgid(%d): %s", gid, err)
	}
	if err := syscall.Setuid(uid); err != nil {
		Fatal.Printf("setuid(%d): %s", uid, err)
	}

	return uid, gid
}
