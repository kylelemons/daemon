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

var signals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGHUP,
	syscall.SIGUSR1,
}

func sigAction(sig os.Signal) int {
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		return sigShutdown
	case syscall.SIGHUP:
		return sigRestart
	case syscall.SIGUSR1:
		return sigStackDump
	}
	return sigUnknown
}
