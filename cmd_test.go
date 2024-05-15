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
	"os"
	"time"

	"kreklow.us/go/cli"
)

func Example() {
	cmd := cli.NewCmd()
	host := cmd.FlagSet.String("host", "localhost", "host name")
	user := cmd.FlagSet.String("user", "", "user name")

	err := cmd.FlagSet.Parse([]string{"-user", "test"})
	if err != nil {
		cmd.Eprintln("unexpected error:", err)
	}

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
		cmd.Printf("connecting: %s@%s\n", *user, *host)

		time.Sleep(time.Second)

		msgs <- []byte("Message")

		cmd.Exit(nil)
	}()

	err = cmd.Wait()
	if err != nil {
		cmd.Eprintln(err)
		os.Exit(1)
	}

	// Output:
	// connecting: test@localhost
	// Message
	// Cleaned up
}
