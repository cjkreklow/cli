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
// The primary design is to help manage the graceful shutdown of
// long-running processes with multiple goroutines.
//
//   func main() {
//     cmd := NewCmd()
//     cmd.AddWait()
//     go process(cmd)
//     cmd.Wait()
//   }
//
//   func process(cmd *Cmd) {
//     defer cmd.Done()
//     exit := cmd.ExitChannel()
//     for {
//       select {
//       case <-exit:
//         break
//       default:
//       }
//       // processing
//     }
//     // cleanup
//   }
//
package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Cmd is the primary structure for maintaining application state. It
// should not be created directly, instead use NewCmd to return a
// properly initialized Cmd.
type Cmd struct {
	wg          *sync.WaitGroup
	debug       *log.Logger
	exitChan    chan bool
	exitOnce    sync.Once
	exitTimeout time.Duration
}

// NewCmd returns a new initialized Cmd configured with default settings.
func NewCmd() *Cmd {
	c := new(Cmd)
	c.wg = new(sync.WaitGroup)
	c.debug = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	c.exitChan = make(chan bool, 1)
	c.exitTimeout = 5 * time.Second
	go c.watchExitSignal()

	return c
}

// D returns the debug logger instance created by NewCmd. It is intended
// to be used directly, as in:
//
//   c.D().Print("debug message")
//
func (c *Cmd) D() *log.Logger {
	return c.debug
}

// SetExitTimeout sets the length of time Exit() waits before forcing
// the application to exit. NewCmd sets a default timeout of 5 seconds.
func (c *Cmd) SetExitTimeout(t time.Duration) {
	c.exitTimeout = t
}

// Exit gracefully closes the application by closing the exit channel
// and waiting for pending operations to finish. In order to be
// successful, the calling application must make use of ExitChannel and
// AddWait.
func (c *Cmd) Exit() {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		c.debug.Printf("exit called in %s line %v", f, l)
	}
	c.exitOnce.Do(func() {
		c.debug.Println("exit triggered")
		close(c.exitChan)
		c.debug.Println("exit channel closed")
		go func() {
			<-time.After(c.exitTimeout)
			fmt.Println("timeout during exit")
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
	var sig os.Signal
	var ok bool

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig, ok = <-sigChan:
		if ok {
			c.debug.Println("caught signal", sig)
		} else {
			c.debug.Println("signal channel closed")
		}
	case <-c.exitChan:
		c.debug.Println("exiting signal watcher")
		ok = true
	}

	c.Exit()

	if ok {
		<-sigChan
		fmt.Println("exit forced")
		os.Exit(1)
	}
}
