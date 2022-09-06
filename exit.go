// Copyright 2020 Collin Kreklow
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
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ExitHandler provides the ability to gracefully shut down an
// application, expanding on the functionality of sync.WaitGroup.
//
// Calling Exit will close the exit channel C, providing the ability for
// any goroutines watching C to perform clean up tasks and return.
//
// Calling Watch with a list of signals will set up a goroutine to
// receive those signals and call Exit, allowing for simple trapping of
// Ctrl-C and kill by passing os.SIGINT and os.SIGTERM.
//
// If a timeout has been set, the closure of the exit channel will also
// trigger a timer which calls os.Exit upon expiration. Sending an exit
// signal during the timeout will abort the timer and call os.Exit
// immediately.
//
// If an error is passed to Exit, the error will be returned to the
// caller of Wait once all the goroutines being awaited call Done. If a
// timeout or signal based forced exit occurs, the error message will be
// printed to os.Stderr before os.Exit is called.
type ExitHandler struct {
	timeout int64 // guarantee 64 bit alignment on 32 bit platforms

	wg sync.WaitGroup

	// C is the exit channel. Must call Add or Watch before attempting
	// to receive from C.
	C <-chan bool

	// ec is the send end of C, reserved for closing.
	ec chan<- bool

	sc chan os.Signal

	exitOnce  sync.Once
	watchOnce sync.Once

	err error
}

// SetTimeout sets the timeout duration. A zero or negative value waits
// indefinitely.
func (e *ExitHandler) SetTimeout(t time.Duration) {
	atomic.StoreInt64(&e.timeout, int64(t))
}

// Exit closes the exit channel and starts the timeout timer, if
// applicable. The error value passed to the first Exit call will be
// passed as the return value of Wait. Exit is safe to call multiple
// times, all calls after the first are ignored.
func (e *ExitHandler) Exit(err error) {
	e.exitOnce.Do(func() {
		e.err = err

		close(e.ec)

		t := atomic.LoadInt64(&e.timeout)

		if t > 0 {
			go e.timeoutWait(t)
		}
	})
}

// timeoutWait implements the timeout, called once by Exit.
func (e *ExitHandler) timeoutWait(t int64) {
	select {
	case <-time.After(time.Duration(t)):
		fmt.Fprintln(os.Stderr, "exit forced by timeout")
	case <-e.sc:
		fmt.Fprintln(os.Stderr, "exit forced by signal")
	}

	if e.err != nil {
		fmt.Fprintln(os.Stderr, e.err)
	}

	os.Exit(int(syscall.ETIME))
}

// Add updates the WaitGroup counter, adding or subtracting as
// appropriate. Add will panic if the counter goes negative.
//
// Add also initializes exit channel C if it has not been initialized
// previously.
func (e *ExitHandler) Add(n int) {
	if e.ec == nil {
		c := make(chan bool)
		e.C = c
		e.ec = c
	}

	e.wg.Add(n)
}

// Done removes one from the WaitGroup counter.
func (e *ExitHandler) Done() {
	e.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero. The return value is
// the first error value passed to Exit.
func (e *ExitHandler) Wait() error {
	e.wg.Wait()

	return e.err
}

// Watch takes a list of signals to receive from the operating system
// which will trigger Exit. Watch can be called multiple times, each
// call to Watch will replace the previous list of signals with the new
// list. An empty list will stop receiving signals from the OS.
func (e *ExitHandler) Watch(signals ...os.Signal) {
	if e.sc == nil {
		e.sc = make(chan os.Signal, 1)
	}

	signal.Stop(e.sc)

	if len(signals) == 0 {
		return
	}

	signal.Notify(e.sc, signals...)

	e.watchOnce.Do(func() {
		if e.ec == nil {
			c := make(chan bool)
			e.C = c
			e.ec = c
		}

		go func() {
			select {
			case <-e.sc:
			case <-e.C:
				return
			}

			e.Exit(nil)
		}()
	})
}
