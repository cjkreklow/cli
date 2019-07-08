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

	"kreklow.us/go/cli"
)

func ExampleCmd_Print() {
	cmd := cli.NewCmd()
	cmd.Print("Hello", 123)

	// Output: Hello123
}

func ExampleCmd_Println() {
	cmd := cli.NewCmd()
	cmd.Println("Countdown", 3, 2, 1)

	// Output: Countdown 3 2 1
}

func ExampleCmd_Printf() {
	cmd := cli.NewCmd()
	cmd.Printf("%s %d = %x", "Convert", 123, 123)

	// Output: Convert 123 = 7b
}

func ExampleCmd_EPrint() {
	cmd := cli.NewCmd()
	cmd.SetErrorWriter(os.Stdout)
	cmd.EPrint("Hello", 123)

	// Output: Hello123
}

func ExampleCmd_EPrintln() {
	cmd := cli.NewCmd()
	cmd.SetErrorWriter(os.Stdout)
	cmd.EPrintln("Countdown", 3, 2, 1)

	// Output: Countdown 3 2 1
}

func ExampleCmd_EPrintf() {
	cmd := cli.NewCmd()
	cmd.SetErrorWriter(os.Stdout)
	cmd.EPrintf("%s %d = %x", "Convert", 123, 123)

	// Output: Convert 123 = 7b
}
