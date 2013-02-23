// +build linux darwin

package daemon

import (
	"flag"
	"os/user"
	"strconv"

	// Used for Setgid and Setuid
	"syscall"
)

// A Privileges stores the desired privileges of a process
// and metadata after they have been dropped.
//
// In the future, this might be extended to also include
// capabilities.
type Privileges struct {
	Username string // User to whom to drop privileges

	UID, GID int // User and Group ID after drop
}

// Drop drops to the configured privileges and returns
// if any dropping was intended.  If dropped privileges
// (that is, a nonzero Username) were requested but
// failed, the process aborts for safety reasons.
func (p *Privileges) Drop() (dropped bool) {
	if p.Username != "" {
		p.UID, p.GID = chuser(p.Username)
		dropped = true
	}
	return dropped
}

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

// PrivilegesFlag registers a flag which, when set, will cause the returned Privileges
// object to drop to the given username.  Recommended default value is "nobody".
func PrivilegesFlag(name, def string) *Privileges {
	p := new(Privileges)
	flag.StringVar(&p.Username, name, def, "User to whom to drop privileges (if set)")
	return p
}
