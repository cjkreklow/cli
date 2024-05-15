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

package cli_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	expect "github.com/Netflix/go-expect"
	"kreklow.us/go/cli"
)

func TestLprintf(t *testing.T) {
	t.Run("Buffer", testLprintfBuffer)
	t.Run("Console", testLprintfConsole)
}

func testLprintfBuffer(t *testing.T) {
	outbuf := new(bytes.Buffer)
	errbuf := new(bytes.Buffer)

	p := cli.NewTermPrinter()
	p.SetStdout(outbuf)
	p.SetStderr(errbuf)

	writeLprintf(p)

	if outbuf.String() != "print 1\nprint 3\nprint 4\nprint 5\nprint 7\nprint 8\n" {
		t.Error("unexpected output", outbuf.String())
	}

	if errbuf.String() != "print 2\nprint 6\nprint 9\n" {
		t.Error("unexpected output", errbuf.String())
	}
}

func testLprintfConsole(t *testing.T) {
	cons, err := expect.NewConsole()
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	var outstr string

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()

		outstr, err = cons.ExpectString("END")
		if err != nil {
			t.Error("unexpected error", err)
		}
	}()

	p := cli.NewTermPrinter()
	p.SetStdout(cons.Tty())
	p.SetStderr(cons.Tty())

	writeLprintf(p)
	p.Print("END")

	wg.Wait()

	if outstr != "print 1\r\nprint 2\r\nprint 3\r\nprint 4\r\n\x1b[1A\x1b[2Kprint 5\r\nprint 6\r\nprint 7\r\nprint 8\r\nprint 9\r\nEND" {
		t.Error("unexpected output", outstr)
	}

	_, err = p.Lprintf("TEST\nTEST\n")
	if err != nil {
		t.Error("unexpected error", err)
	}

	err = cons.Tty().Close()
	if err != nil {
		t.Error("unexpected error", err)
	}

	defer func() {
		p := recover()

		err, ok := p.(error)
		if !ok ||
			!strings.HasPrefix(err.Error(), "write") ||
			!strings.HasSuffix(err.Error(), "file already closed") {
			t.Error("unexpected panic:", p)
		}
	}()

	_, err = p.Lprintf("TEST\n")
	t.Error("expected panic, got", err)
}

func writeLprintf(p *cli.TermPrinter) {
	p.Print("print 1\n")
	p.Eprintf("print %d\n", 2)
	p.Println("print 3")
	p.Lprintf("print %d\n", 4)
	p.Lprintf("print %d\n", 5)
	p.Eprint("print 6\n")
	p.Lprintf("print %d\n", 7)
	p.Printf("print %d\n", 8)
	p.Eprintln("print 9")
}
