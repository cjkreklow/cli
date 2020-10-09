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
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/mattn/go-isatty"
)

// lockingWriter is a simple mutex-protected writer.
type lockingWriter struct {
	m sync.Mutex
	w io.Writer
}

// Write passes the provided data to the embedded io.Writer.
func (lw *lockingWriter) Write(b []byte) (n int, err error) {
	lw.m.Lock()
	n, err = lw.w.Write(b)
	lw.m.Unlock()

	return
}

// TermPrinter provides printing functions based on a standard unix
// terminal. The Print* functions direct output to os.Stdout, while the
// Eprint* functions direct output to os.Stderr.
//
// TermPrinter also includes a "live printing" function in Lprintf which
// provides a non-scrolling output for continuously-updating status
// information. Messages printed via Print* or Eprint* will not be
// overwritten by Lprintf.
//
// TermPrinter provides locking over the output writers, so it is safe
// to call concurrently from multiple goroutines.
//
// If TermPrinter is not created with NewTermPrinter, SetStdout and
// SetStderr must be called before use.
//
type TermPrinter struct {
	livecount uint32

	outIsTerm bool
	errIsTerm bool

	out io.Writer
	err io.Writer

	livebuf bytes.Buffer
}

// NewTermPrinter returns a TermPrinter set to output to os.Stdout and
// os.Stderr.
func NewTermPrinter() *TermPrinter {
	return &TermPrinter{
		out: &lockingWriter{w: os.Stdout},
		err: &lockingWriter{w: os.Stderr},
	}
}

// SetStdout sets the destination for calls to Print, Printf, Println
// and Lprintf.
func (tp *TermPrinter) SetStdout(w io.Writer) {
	tp.out = &lockingWriter{w: w}
	tp.outIsTerm = false

	if f, ok := w.(*os.File); ok {
		tp.outIsTerm = isatty.IsTerminal(f.Fd())
	}
}

// SetStderr sets the destination for calls to EPrint, EPrintf and
// EPrintln.
func (tp *TermPrinter) SetStderr(w io.Writer) {
	tp.err = &lockingWriter{w: w}
	tp.errIsTerm = false

	if f, ok := w.(*os.File); ok {
		tp.errIsTerm = isatty.IsTerminal(f.Fd())
	}
}

// Print operates in the manner of fmt.Print, writing to Stdout.
func (tp *TermPrinter) Print(v ...interface{}) (int, error) {
	if tp.outIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprint(tp.out, v...)
}

// Printf operates in the manner of fmt.Printf, writing to Stdout.
func (tp *TermPrinter) Printf(f string, v ...interface{}) (int, error) {
	if tp.outIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprintf(tp.out, f, v...)
}

// Println operates in the manner of fmt.Println, writing to Stdout.
func (tp *TermPrinter) Println(v ...interface{}) (int, error) {
	if tp.outIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprintln(tp.out, v...)
}

// Lprintf implements a "live update" version of fmt.Printf. If Stdout
// appears to be a terminal, the previously output line(s) will be
// cleared before the new line(s) are written.
//
// While Lprintf is safe for concurrent use with Print* and Eprint*,
// concurrent use of Lprintf will conflict, overwriting the previous
// output.
func (tp *TermPrinter) Lprintf(f string, v ...interface{}) (int, error) {
	if !tp.outIsTerm {
		return fmt.Fprintf(tp.out, f, v...)
	}

	tp.clearLiveLines()
	tp.livebuf.Reset()

	fmt.Fprintf(&tp.livebuf, f, v...)

	b := tp.livebuf.Bytes()

	atomic.StoreUint32(&tp.livecount, uint32(bytes.Count(b, []byte{'\n'})))

	return tp.out.Write(b)
}

// Eprint operates in the manner of fmt.Print, writing to Stderr.
func (tp *TermPrinter) Eprint(v ...interface{}) (int, error) {
	if tp.errIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprint(tp.err, v...)
}

// Eprintf operates in the manner of fmt.Printf, writing to Stderr.
func (tp *TermPrinter) Eprintf(f string, v ...interface{}) (int, error) {
	if tp.errIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprintf(tp.err, f, v...)
}

// Eprintln operates in the manner of fmt.Println, writing to Stderr.
func (tp *TermPrinter) Eprintln(v ...interface{}) (int, error) {
	if tp.errIsTerm {
		tp.resetLiveLines()
	}

	return fmt.Fprintln(tp.err, v...)
}

func (tp *TermPrinter) resetLiveLines() {
	atomic.StoreUint32(&tp.livecount, 0)
}

//nolint:gochecknoglobals // improves performance of clearLiveLines
var clearline = []byte("\x1b[1A\x1b[2K")

func (tp *TermPrinter) clearLiveLines() {
	ll := atomic.LoadUint32(&tp.livecount)

	for l := uint32(0); l < ll; l++ {
		_, err := tp.out.Write(clearline)
		if err != nil {
			panic(err)
		}
	}

	tp.resetLiveLines()
}
