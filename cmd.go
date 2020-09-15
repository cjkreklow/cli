// Copyright 2019 Collin Kreklow
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Cmd is the primary structure for maintaining application state. It
// should not be created directly, instead use NewCmd to return a
// properly initialized Cmd.
type Cmd struct {
	flagSet      *flag.FlagSet
	outWriter    io.Writer
	outLock      sync.Mutex
	errWriter    io.Writer
	errLock      sync.Mutex
	outLiveBuf   bytes.Buffer
	outLiveLines int
	exitTimeout  atomic.Value
	exitWg       *sync.WaitGroup
	exitChan     chan bool
	exitOnce     sync.Once
	errIsTerm    bool
	outIsTerm    bool
	err          error
}

// NewCmd returns a new initialized Cmd configured with default settings.
func NewCmd() *Cmd {
	c := new(Cmd)
	c.exitWg = new(sync.WaitGroup)
	c.exitChan = make(chan bool, 1)

	c.SetExitTimeout(5 * time.Second)
	c.SetOutputWriter(os.Stdout)
	c.SetErrorWriter(os.Stderr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go c.watchExitSignal(sigChan)

	c.flagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	return c
}

// Flags returns an embedded FlagSet.
func (c *Cmd) Flags() *flag.FlagSet {
	return c.flagSet
}
