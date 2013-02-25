// echo runs a simple echo server
package main

import (
	"flag"
	"io"
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
