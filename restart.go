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

package daemon

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"
)

// Only allow one routine to try to stop the binary
var stopOnce = make(chan bool, 1)

func init() {
	stopOnce <- true
}

func copyFlags() (cmd *exec.Cmd, ports []*WaitListener) {
	cmd = exec.Command(os.Args[0])

	flag.VisitAll(func(f *flag.Flag) {
		switch val := f.Value.(type) {
		case *listenFlag:
			if val.listener == nil {
				// flag hasn't been listened yet, so just pass through
				break
			}

			// The extra files list doesn't include stdin/out/err
			fd := 3 + len(cmd.ExtraFiles)

			// Add this flag to the cmd
			cmd.Args = append(cmd.Args, fmt.Sprintf("--%s=&%d", f.Name, fd))
			cmd.ExtraFiles = append(cmd.ExtraFiles, val.listener.File())

			// return the port so it can be closed
			ports = append(ports, val.listener)
			return
		case *forkFlag:
			// Don't pass fork on to subprocesses
			return
		}
		cmd.Args = append(cmd.Args, fmt.Sprintf("--%s=%s", f.Name, f.Value))
	})
	return
}

func spawn(cmd *exec.Cmd) {
	Verbose.Printf("Spawning process: %q %q", cmd.Args[0], cmd.Args[1:])
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		Fatal.Printf("Exec failed: %s", err)
	}
}

// Restart re-execs the current process, passing all of the same flags,
// except that ListenFlags will be replaced with "&fd" to copy the file
// descriptor from this process.  Restart does not return.
func Restart(timeout time.Duration) {
	<-stopOnce

	cmd, ports := copyFlags()
	for _, w := range ports {
		w.Stop()
		// Send noop connections to free up the accept loops
		w.noop()
	}
	spawn(cmd)

	// Wait for all connections to close out
	done := make(chan bool)
	go func() {
		defer close(done)
		for _, w := range ports {
			w.Wait()
		}
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		Fatal.Printf("Restart timed out after %s", timeout)
	}
	Verbose.Printf("Restart complete")
	os.Exit(0)
}

// Shutdown closes all ListenFlags and waits for their connections to
// finish.  Shutdown does not return.
func Shutdown(timeout time.Duration) {
	<-stopOnce

	_, ports := copyFlags()
	for _, w := range ports {
		w.Close()
	}

	// Wait for all connections to close out
	done := make(chan bool)
	go func() {
		defer close(done)
		for _, w := range ports {
			w.Wait()
		}
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		Fatal.Printf("Shutdown timed out after %s", timeout)
	}
	Info.Printf("Shutdown complete")
	os.Exit(0)
}

// A Forker knows how to duplicate the main process by replicating its flags.
// Fork only returns in the subprocess.  The parent process exits, and the
// child process writes its pid to the pidfile.
type Forker interface {
	Fork()
}

type forkFlag struct {
	fork    bool
	pidfile string
}

func (f *forkFlag) String() string {
	return fmt.Sprintf("%v", f.fork)
}

func (f *forkFlag) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	f.fork = b
	return nil
}

func (f *forkFlag) Fork() {
	if f.fork {
		<-stopOnce

		// Don't fork in the child
		f.fork = false

		Verbose.Printf("Forking into the background")
		cmd, _ := copyFlags()
		spawn(cmd)
		os.Exit(0)
	}

	pidfile, err := os.Create(f.pidfile)
	if err != nil {
		Error.Printf("Failed to create pidfile: %s", err)
		return
	}
	defer pidfile.Close()

	fmt.Fprintf(pidfile, "%d\n", os.Getpid())
	Verbose.Printf("Wrote PID to %s", f.pidfile)
}

// ForkPIDFlags registers two flags, with the given names, and returns a Forker
// which should be called to manage forking and writing the PID to file.
func ForkPIDFlags(forkFlagName, pidFlagName string, defPIDFile string) Forker {
	f := &forkFlag{}
	flag.StringVar(&f.pidfile, pidFlagName, defPIDFile, "File to which to write PID")
	flag.BoolVar(&f.fork, forkFlagName, false, "Fork into the background")
	return f
}

// LameDuck specifies the duration of the lame duck mode after the
// listener is closed before the binary exits.
var LameDuck = 15 * time.Second

// Run is the last thing to call from main.  It does not return.
//
// Run handles the following signals:
//   SIGINT    - Calls Shutdown
//   SIGTERM   - Calls Shutdown
//   SIGHUP    - Calls Restart
//   SIGUSR1   - Dumps a stack trace to the logs
//
// If another signal is received during Shutdown or Restart, the process
// will terminate immediately.
func Run() {
	incoming := make(chan os.Signal, 10)
	signal.Notify(incoming, signals...)
	for sig := range incoming {
		select {
		case <-stopOnce:
			stopOnce <- true
		default:
			Fatal.Printf("Aborted by signal during shutdown")
		}

		switch sigAction(sig) {
		case sigShutdown:
			go Shutdown(LameDuck)
		case sigRestart:
			go Restart(LameDuck)
		case sigStackDump:
			V(-5).Printf("Stack dump:\n" + stack())
		default:
			Warning.Printf("Unknown signal: %s", sig)
		}
	}
}

// Return values for platform-specific sigAction
const (
	sigUnknown = iota
	sigShutdown
	sigRestart
	sigStackDump
)
