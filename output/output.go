// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package output

import (
	"context"
	"io"
	"os"
)

func WithOutput(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, outKey{}, w)
}

func WithErr(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, errKey{}, w)
}

func FromContext(ctx context.Context) io.Writer {
	w, ok := ctx.Value(outKey{}).(io.Writer)
	if !ok {
		return os.Stdout
	}
	return w
}

func ErrFromContext(ctx context.Context) io.Writer {
	w, ok := ctx.Value(errKey{}).(io.Writer)
	if !ok {
		return os.Stderr
	}
	return w
}

type outKey struct{}

type errKey struct{}
