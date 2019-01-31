// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package prefixio_test

import (
	"io"
	"strings"
	"testing"

	"github.com/szabba/assert"

	"github.com/szabba/unibuild/prefixio"
)

func TestWriterEmpty(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := w.Write(nil)

	// then
	assert.That(n == 0, t.Errorf, "n: got %d, want %d", n, 0)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "", t.Errorf, "out: got %q, want %q", out.String(), "")
}

func TestWriterOneLine(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := io.WriteString(w, "a")

	// then
	assert.That(n == 1, t.Errorf, "n: got %d, want %d", n, 1)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> a", t.Errorf, "out: got %q, want %q", out.String(), "> a")
}

func TestWriterOneLineWithNewline(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := io.WriteString(w, "a\n")

	// then
	assert.That(n == 2, t.Errorf, "n: got %d, want %d", n, 2)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> a\n", t.Errorf, "out: got %q, want %q", out.String(), "> a\n")
}

func TestWriteTwoChunksWithNewlineBetween(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := io.WriteString(w, "a\nb")

	// then
	assert.That(n == 3, t.Errorf, "n: got %d, want %d", n, 3)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> a\n> b", t.Errorf, "out: got %q, want %q", out.String(), "> a\n> b")
}

func TestWriteNewline(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := io.WriteString(w, "\n")

	// then
	assert.That(n == 1, t.Errorf, "n: got %d, want %d", n, 1)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> \n", t.Errorf, "out: got %q, want %q", out.String(), "> \n")
}

func TestWriteTwoNewlines(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	// when
	n, err := io.WriteString(w, "\n\n")

	// then
	assert.That(n == 2, t.Errorf, "n: got %d, want %d", n, 2)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> \n> \n", t.Errorf, "out: got %q, want %q", out.String(), "> \n> \n")
}

func TestWriteMultipleLineChunks(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	n, err := io.WriteString(w, "a")
	assert.That(n == 1, t.Fatalf, "n: got %d, want %d", n, 1)
	assert.That(err == nil, t.Fatalf, "unexpected error: %s", err)

	// when
	n, err = io.WriteString(w, "b")

	// then
	assert.That(n == 1, t.Errorf, "n: got %d, want %d", n, 1)
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
	assert.That(out.String() == "> ab", t.Errorf, "out: got %q, want %q", out.String(), "> ab")
}

func TestWriteMultipleLineChunksAndNewlines(t *testing.T) {
	// given
	out := new(strings.Builder)
	w := prefixio.NewWriter(out, "> ")

	n, err := io.WriteString(w, "a")
	assert.That(n == 1, t.Fatalf, "n: got %d, want %d", n, 1)
	assert.That(err == nil, t.Fatalf, "unexpected error: %s", err)

	// when
	n, err = io.WriteString(w, "b\nc")

	// then
	assert.That(n == 3, t.Fatalf, "n: got %d, want %d", n, 3)
	assert.That(err == nil, t.Fatalf, "unexpected error: %s", err)
	assert.That(out.String() == "> ab\n> c", t.Errorf, "out: got %q, want %q", out.String(), "> ab\n> c")
}
