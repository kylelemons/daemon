// +build linux darwin

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
	"os"
	"syscall"
)

// RedirectStdout will cause anything written to standard output to be also
// written to the LogFileFlagged file.  In particular, when this is true, panic
// traces and standard uses of the "log" package will find their way into the
// logfile.  Set this to false during init to suppress this behavior.
var RedirectStdout = true

func redirectStdout() {
	if !RedirectStdout {
		return
	}

	syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
}
