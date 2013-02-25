package daemon

import (
	"flag"
)

// A Privileges stores the desired privileges of a process
// and metadata after they have been dropped.
//
// In the future, this might be extended to also include
// capabilities.
type Privileges struct {
	Username string // User to whom to drop privileges
}

// Drop drops to the configured privileges and returns
// if any dropping was intended.  If dropped privileges
// (that is, a nonzero Username) were requested but
// failed, the process aborts for safety reasons.
func (p *Privileges) Drop() (dropped bool) {
	if p.Username != "" {
		chuser(p.Username)
		dropped = true
	}
	return dropped
}

// PrivilegesFlag registers a flag which, when set, will cause the returned Privileges
// object to drop to the given username.  Recommended default value is "nobody".
func PrivilegesFlag(name, def string) *Privileges {
	p := new(Privileges)
	flag.StringVar(&p.Username, name, def, "User to whom to drop privileges (if set)")
	return p
}
