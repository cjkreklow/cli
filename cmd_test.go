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

package cli_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"kreklow.us/go/cli"
)

//nolint:goerr113 // keep example simple
func Example() {
	cmd := cli.NewCmd()
	msgs := make(chan []byte, 1)

	cmd.Add(1)

	// message receiver
	go func() {
		err := errors.New("unexpected shutdown")

		defer cmd.Done()    // deferring Done() and Exit() helps ensure a clean
		defer cmd.Exit(err) // shutdown if the goroutine returns unexpectedly

	loop:
		for {
			select {
			case <-cmd.C:
				// exit signal, go to cleanup
				break loop
			case m := <-msgs:
				// processing tasks
				cmd.Printf("%s\n", m)
			}
		}

		// cleanup tasks
		cmd.Println("Cleaned up")
	}()

	// message sender
	go func() {
		time.Sleep(time.Second)

		msgs <- []byte("Message")

		cmd.Exit(nil)
	}()

	err := cmd.Wait()
	if err != nil {
		cmd.Eprintln(err)
		os.Exit(1)
	}

	// Output:
	// Message
	// Cleaned up
}

func TestFlagSet(t *testing.T) {
	cmd := cli.NewCmd()
	host := cmd.Flags().String("host", "localhost", "host name")
	user := cmd.Flags().String("user", "", "user name")

	err := cmd.Flags().Parse([]string{"-user", "test"})
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if *host != "localhost" {
		t.Error("expected: localhost  received: ", host)
	}

	if *user != "test" {
		t.Error("expected: test  received: ", user)
	}
}
