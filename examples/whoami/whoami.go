// whoami prints out the username after dropping privileges.
package main

import (
	"fmt"
	"os/user"

	"flag"
	"kylelemons.net/go/daemon"
)

var (
	privs = daemon.PrivilegesFlag("user", "nobody")
)

func main() {
	flag.Parse()
	privs.Drop()

	user, err := user.Current()
	if err != nil {
		daemon.Fatal.Printf("current: %s", err)
	}

	fmt.Println(user.Username)
}
