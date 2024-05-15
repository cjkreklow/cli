// Copyright 2024 Collin Kreklow
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

//nolint:goerr113 // keep examples simple
package cli_test

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"kreklow.us/go/cli"
)

func ExampleExitHandler() {
	eh := new(cli.ExitHandler)
	msgs := make(chan []byte, 1)

	eh.Add(1)

	// message receiver
	go func() {
		err := errors.New("unexpected shutdown")

		defer eh.Done()    // deferring Done() and Exit() helps ensure a clean
		defer eh.Exit(err) // shutdown if the goroutine returns unexpectedly

	loop:
		for {
			select {
			case <-eh.C:
				// exit signal, go to cleanup
				break loop
			case m := <-msgs:
				// do some work
				fmt.Printf("%s\n", m)
			}
		}

		// cleanup tasks
		fmt.Println("Cleaned up")
	}()

	// message sender
	go func() {
		msgs <- []byte("Message")

		time.Sleep(50 * time.Millisecond) // do some work

		eh.Exit(nil)
	}()

	err := eh.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Output:
	// Message
	// Cleaned up
}

func TestSignalExit(t *testing.T) {
	t.Run("Normal", testExitSignal)
	t.Run("Reset", testExitReset)
	t.Run("None", testExitNone)
}

func testExitSignal(t *testing.T) {
	eh := new(cli.ExitHandler)

	eh.Watch(syscall.SIGHUP)

	eh.Add(1)

	go func() {
		<-eh.C
		eh.Done()
	}()

	err := syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	err = eh.Wait()
	if err != nil {
		t.Error("unexpected error:", err)
	}

	signal.Reset()
}

func testExitReset(t *testing.T) {
	eh := new(cli.ExitHandler)

	eh.Watch(syscall.SIGHUP)

	eh.Add(1)

	go func() {
		<-eh.C
		eh.Done()
	}()

	eh.Watch(syscall.SIGUSR1)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	err = eh.Wait()
	if err != nil {
		t.Error("unexpected error:", err)
	}

	signal.Reset()
}

func testExitNone(t *testing.T) {
	eh := new(cli.ExitHandler)

	eh.Watch(syscall.SIGUSR1)
	eh.SetTimeout(10 * time.Second)

	eh.Add(1)

	go func() {
		<-eh.C
		time.Sleep(time.Second)
		eh.Done()
	}()

	eh.Watch()

	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	go func() {
		time.Sleep(time.Second)
		eh.Exit(errors.New("testing error")) //nolint:goerr113 // ignore in test
	}()

	err = eh.Wait()
	if err == nil {
		t.Error("expected error, received nil")
	} else if err.Error() != "testing error" {
		t.Error("unexpected error:", err)
	}

	signal.Reset()
}
