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
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// SetOutputWriter sets the destination for calls to Print(), Printf()
// and Println(). NewCmd() sets the default output writer is os.Stdout.
func (c *Cmd) SetOutputWriter(w io.Writer) {
	c.outLock.Lock()
	defer c.outLock.Unlock()
	c.outWriter = w
	c.outIsTerm = false

	if f, ok := w.(*os.File); ok {
		c.outIsTerm = isatty.IsTerminal(f.Fd())
	}
}

// Print is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Print(v ...interface{}) (int, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if c.outIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprint(c.outWriter, v...)
}

// Printf is a wrapper around fmt.Printf with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Printf(f string, v ...interface{}) (int, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if c.outIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprintf(c.outWriter, f, v...)
}

// Lprintf implements a "live update" version of cmd.Printf. If the
// io.Writer specified by SetOutputWriter appears to be a terminal, the
// previously output line(s) will be cleared before the new line(s) are
// written. It is safe for concurrent use, although concurrent updates
// will overwrite each other.
func (c *Cmd) Lprintf(f string, v ...interface{}) (int, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if !c.outIsTerm {
		return fmt.Fprintf(c.outWriter, f, v...)
	}

	c.clearLiveLines()
	c.outLiveBuf.Reset()
	fmt.Fprintf(&c.outLiveBuf, f, v...)
	b := c.outLiveBuf.Bytes()
	c.outLiveLines = bytes.Count(b, []byte{'\n'})

	return c.outWriter.Write(b)
}

// Println is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Println(v ...interface{}) (int, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if c.outIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprintln(c.outWriter, v...)
}

// SetErrorWriter sets the destination for calls to EPrint(), EPrintf()
// and EPrintln(). NewCmd() sets the default error writer is os.Stderr.
func (c *Cmd) SetErrorWriter(w io.Writer) {
	c.errLock.Lock()
	defer c.errLock.Unlock()

	c.errWriter = w
	c.errIsTerm = false

	if f, ok := w.(*os.File); ok {
		c.errIsTerm = isatty.IsTerminal(f.Fd())
	}
}

// Eprint is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetErrorWriter. It is safe for concurrent use.
func (c *Cmd) Eprint(v ...interface{}) (int, error) {
	c.errLock.Lock()
	defer c.errLock.Unlock()

	if c.errIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprint(c.errWriter, v...)
}

// Eprintf is a wrapper around fmt.Printf with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) Eprintf(f string, v ...interface{}) (int, error) {
	c.errLock.Lock()
	defer c.errLock.Unlock()

	if c.errIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprintf(c.errWriter, f, v...)
}

// Eprintln is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) Eprintln(v ...interface{}) (int, error) {
	c.errLock.Lock()
	defer c.errLock.Unlock()

	if c.errIsTerm {
		c.outLiveLines = 0
	}

	return fmt.Fprintln(c.errWriter, v...)
}

func (c *Cmd) clearLiveLines() {
	for l := 0; l < c.outLiveLines; l++ {
		_, err := c.outWriter.Write([]byte("\x1b[1A\x1b[2K"))
		if err != nil {
			panic(err)
		}
	}

	c.outLiveLines = 0
}
