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

// Package cli provides simple framework for command line applications.
// The primary goal is to help manage the graceful shutdown of
// long-running processes with multiple goroutines.
//
package cli

import (
	"fmt"
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
	wg          *sync.WaitGroup
	exitChan    chan bool
	exitOnce    sync.Once
	exitTimeout atomic.Value
	outWriter   io.Writer
	outLock     sync.Mutex
	errWriter   io.Writer
	errLock     sync.Mutex
}

// NewCmd returns a new initialized Cmd configured with default settings.
func NewCmd() *Cmd {
	c := new(Cmd)
	c.wg = new(sync.WaitGroup)
	c.exitChan = make(chan bool, 1)

	c.SetExitTimeout(5 * time.Second)
	c.SetOutputWriter(os.Stdout)
	c.SetErrorWriter(os.Stderr)

	go c.watchExitSignal()

	return c
}

// SetOutputWriter sets the destination for calls to Print(), Printf()
// and Println(). NewCmd() sets the default output writer is os.Stdout.
func (c *Cmd) SetOutputWriter(w io.Writer) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	c.outWriter = w
}

// Print is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Print(v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprint(c.outWriter, v...)
}

// Printf is a wrapper around fmt.Printf with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Printf(f string, v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprintf(c.outWriter, f, v...)
}

// Println is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Println(v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprintln(c.outWriter, v...)
}

// SetErrorWriter sets the destination for calls to EPrint(), EPrintf()
// and EPrintln(). NewCmd() sets the default error writer is os.Stderr.
func (c *Cmd) SetErrorWriter(w io.Writer) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	c.errWriter = w
}

// EPrint is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetErrorWriter. It is safe for concurrent use.
func (c *Cmd) EPrint(v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprint(c.errWriter, v...)
}

// EPrintf is a wrapper around fmt.Printf with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) EPrintf(f string, v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprintf(c.errWriter, f, v...)
}

// EPrintln is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) EPrintln(v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprintln(c.errWriter, v...)
}

// SetExitTimeout sets the length of time Exit() waits before forcing
// the application to exit. NewCmd sets a default timeout of 5 seconds.
func (c *Cmd) SetExitTimeout(t time.Duration) {
	c.exitTimeout.Store(t)
}

// Exit gracefully closes the application by closing the exit channel
// and waiting for pending operations to finish. In order to be
// successful, the calling application must make use of ExitChannel and
// AddWait.
func (c *Cmd) Exit() {
	c.exitOnce.Do(func() {
		close(c.exitChan)
		go func() {
			<-time.After(c.exitTimeout.Load().(time.Duration))
			c.EPrintln("exit forced by timeout")
			os.Exit(1)
		}()
	})
}

// ExitChannel returns a read-only channel that will be closed when the
// application is exiting.
func (c *Cmd) ExitChannel() <-chan bool {
	return c.exitChan
}

// AddWait notifies Cmd that a goroutine is running and Wait() should
// not return until the goroutine is finished. The goroutine must call
// Done() before exiting.
func (c *Cmd) AddWait() {
	c.wg.Add(1)
}

// Done notifies the Cmd that a goroutine previously added with
// AddWait() is now complete.
func (c *Cmd) Done() {
	c.wg.Done()
}

// Wait blocks until all goroutines added with AddWait() have called
// Done(). This should typically be called at the end of main() to avoid
// exiting prematurely.
func (c *Cmd) Wait() {
	c.wg.Wait()
}

// watchExitSignal is an internal function to watch for common keyboard
// interrupt signals and gracefully exit the application.
func (c *Cmd) watchExitSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
	case <-c.exitChan:
	}

	c.Exit()
	<-sigChan
	c.EPrintln("exit forced by signal")
	os.Exit(1)
}
