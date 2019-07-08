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
// The primary goal is to help manage the graceful shutdown of
// long-running processes with multiple goroutines.
//
package cli

import (
	"fmt"
	"io"
)

// SetOutputWriter sets the destination for calls to Print(), Printf()
// and Println(). NewCmd() sets the default output writer is os.Stdout.
func (c *Cmd) SetOutputWriter(w io.Writer) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	c.outWriter = w
}

// Print is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Print(v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprint(c.outWriter, v...)
}

// Printf is a wrapper around fmt.Printf with the destination set to the
// io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Printf(f string, v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprintf(c.outWriter, f, v...)
}

// Println is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetOutputWriter. It is safe for concurrent
// use.
func (c *Cmd) Println(v ...interface{}) (int, error) {
	defer c.outLock.Unlock()
	c.outLock.Lock()
	return fmt.Fprintln(c.outWriter, v...)
}

// SetErrorWriter sets the destination for calls to EPrint(), EPrintf()
// and EPrintln(). NewCmd() sets the default error writer is os.Stderr.
func (c *Cmd) SetErrorWriter(w io.Writer) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	c.errWriter = w
}

// EPrint is a wrapper around fmt.Print with the destination set to the
// io.Writer specified by SetErrorWriter. It is safe for concurrent use.
func (c *Cmd) EPrint(v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprint(c.errWriter, v...)
}

// EPrintf is a wrapper around fmt.Printf with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) EPrintf(f string, v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprintf(c.errWriter, f, v...)
}

// EPrintln is a wrapper around fmt.Println with the destination set to
// the io.Writer specified by SetErrorWriter. It is safe for concurrent
// use.
func (c *Cmd) EPrintln(v ...interface{}) (int, error) {
	defer c.errLock.Unlock()
	c.errLock.Lock()
	return fmt.Fprintln(c.errWriter, v...)
}
