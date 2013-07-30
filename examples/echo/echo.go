// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// echo runs a simple echo server
package main

import (
	"flag"
	"io"
	"os"
	"time"

	"kylelemons.net/go/daemon"
)

var (
	echo  = daemon.ListenFlag("echo", "tcp", ":12112", "echo")
	fork  = daemon.ForkPIDFlags("fork", "pidfile", "echo.pid")
	delay = flag.Duration("delay", 1*time.Minute, "Delay between restarts")
)

func main() {
	flag.Parse()
	daemon.Info.Printf("Command-line: %q", os.Args)

	daemon.LogLevel = daemon.Verbose
	fork.Fork()

	port, err := echo.Listen()
	if err != nil {
		daemon.Fatal.Printf("listen: %s", err)
	}

	go func() {
		for {
			conn, err := port.Accept()
			if err == daemon.ErrStopped {
				break
			}
			if err != nil {
				daemon.Error.Printf("accept: %s", err)
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
		daemon.Info.Printf("Serve loop exited")
	}()

	go func() {
		time.Sleep(*delay)
		daemon.Restart(15 * time.Second)
	}()
	daemon.Run()
}
