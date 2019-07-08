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

package cli_test

import (
	"os"
	"testing"
	"time"

	"kreklow.us/go/cli"
)

func Example() {
	cmd := cli.NewCmd()

	msgs := make(chan []byte, 1)

	cmd.AddWait()
	go func() {
		defer cmd.Done() // deferring Done() and Exit() helps ensure a clean
		defer cmd.Exit() // shutdown if the goroutine returns unexpectedly
	loop:
		for {
			select {
			case <-cmd.ExitChannel():
				// exit signal, go to cleanup
				break loop
			case m := <-msgs:
				// processing tasks
				cmd.Printf("%s", m)
			}
		}
		// cleanup tasks
	}()

	// simulate a message sender
	go func() {
		time.Sleep(time.Second)
		msgs <- []byte("Message")
		cmd.Exit()
	}()

	cmd.Wait()

	// Output: Message
}

func TestFlagSet(t *testing.T) {
	cmd := cli.NewCmd()
	str := cmd.Flags().String("host", "localhost", "host name")
	cmd.Flags().Bool("test.v", false, "verbose output")
	err := cmd.Flags().Parse(os.Args[1:])
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if *str != "localhost" {
		t.Error("expected: localhost  received: ", str)
	}
}
