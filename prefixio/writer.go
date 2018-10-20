// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package prefixio implements code for doing I/O with line prefixes.
package prefixio

import (
	"io"
	"regexp"
)

// Writer implements line-prefixing for an io.Writer object.
type Writer struct {
	dst        io.Writer
	linePrefix []byte
	continued  bool
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
	lineChunk   = regexp.MustCompile(`[^\r\n]*(\r?\n|$)`)
	lineWithEnd = regexp.MustCompile(`[^\r\n]*\r?\n`)
)

func (w *Writer) Write(p []byte) (n int, err error) {
	chunks := lineChunk.FindAll(p, -1)
	for _, c := range chunks {
		dn, err := w.writeChunk(c)
		n += dn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (w *Writer) writeChunk(c []byte) (int, error) {
	if len(c) == 0 {
		return 0, nil
	}

	err := w.writeChunkPrefix()
	if err != nil {
		return 0, err
	}

	w.rememberEnding(c)
	return w.dst.Write(c)
}

func (w *Writer) writeChunkPrefix() error {
	if w.continued {
		return nil
	}
	_, err := w.dst.Write(w.linePrefix)
	return err
}

func (w *Writer) rememberEnding(c []byte) {
	w.continued = !lineWithEnd.Match(c)
}
