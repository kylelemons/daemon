package daemon

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func copyFlags() (arg0 string, flags []string, ports []*WaitListener) {
	arg0 = os.Args[0]
	flag.VisitAll(func(f *flag.Flag) {
		if lf, ok := f.Value.(*listenFlag); ok && lf.listener != nil {
			fd := lf.listener.Dup()
			ports = append(ports, lf.listener)
			flags = append(flags, fmt.Sprintf("--%s=&%d", f.Name, fd))
			return
		}
		flags = append(flags, fmt.Sprintf("--%s=%s", f.Name, f.Value))
	})
	return
}

func spawn(arg0 string, flags []string) {
	Verbose.Printf("Spawning process: %q %q", arg0, flags)
	cmd := exec.Command(arg0, flags...)
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
	arg0, flags, ports := copyFlags()
	for _, w := range ports {
		w.Stop()
		// Send noop connections to free up the accept loops
		w.noop()
	}

	spawn(arg0, flags)

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
	_, _, ports := copyFlags()
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
		switch sigAction(sig) {
		case sigShutdown:
			go Shutdown(LameDuck)
			<-incoming
			Fatal.Printf("Shutdown aborted")
		case sigRestart:
			go Restart(LameDuck)
			<-incoming
			Fatal.Printf("Restart aborted")
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
