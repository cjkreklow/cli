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
	"os"
	"os/signal"
	"syscall"
	"time"
)

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
	c.exitWg.Add(1)
}

// Done notifies the Cmd that a goroutine previously added with
// AddWait() is now complete.
func (c *Cmd) Done() {
	c.exitWg.Done()
}

// Wait blocks until all goroutines added with AddWait() have called
// Done(). This should typically be called at the end of main() to avoid
// exiting prematurely.
func (c *Cmd) Wait() {
	c.exitWg.Wait()
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
