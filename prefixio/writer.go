// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package prefixio implements code for doing I/O with line prefixes.
package prefixio

import (
	"bytes"
	"io"
	"regexp"
)

// Writer implements line-prefixing for an io.Writer object.
type Writer struct {
	dst        io.Writer
	linePrefix []byte

	buf       bytes.Buffer
	keeps     [][2]int
	continued bool
}

// NewWriter returns a new Writer with the given line prefix.
func NewWriter(dst io.Writer, linePrefix string) *Writer {
	return &Writer{
		dst:        dst,
		linePrefix: []byte(linePrefix),
	}
}

var _ io.Writer = new(Writer)

var (
	_LineChunk   = regexp.MustCompile(`[^\r\n]*(\r?\n|$)`)
	_LineWithEnd = regexp.MustCompile(`[^\r\n]*\r?\n`)
)

func (w *Writer) Write(p []byte) (int, error) {
	w.reset()
	w.bufferOutput(p)
	return w.copy()
}

func (w *Writer) reset() {
	w.buf.Reset()
	w.keeps = w.keeps[:0]
}

func (w *Writer) bufferOutput(p []byte) {
	if len(p) == 0 {
		return
	}
	chunks := _LineChunk.FindAll(p, -1)
	for _, c := range chunks {
		w.bufferChunk(c)
	}
}

func (w *Writer) copy() (int, error) {
	n, err := io.Copy(w.dst, &w.buf)
	return w.translateOffset(int(n)), err
}

func (w *Writer) bufferChunk(c []byte) {
	if !w.continued {
		w.buf.Write(w.linePrefix)
	}
	w.buf.Write(c)
	w.keeps = append(w.keeps, w.keep(c))
	w.continued = !_LineWithEnd.Match(c)
}

func (w *Writer) keep(c []byte) [2]int {
	off := 0
	if !w.continued {
		off += len(w.linePrefix)
	}
	if len(w.keeps) != 0 {
		last := len(w.keeps) - 1
		off += w.keeps[last][1]
	}
	return [2]int{off, off + len(c)}
}

func (w *Writer) translateOffset(n int) int {
	out := 0
	for _, keep := range w.keeps {
		from, upto := keep[0], keep[1]
		if upto <= n {
			out += upto - from
		} else if from < upto {
			out += n - from
		}
	}
	return out
}
