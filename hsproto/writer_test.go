// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hsproto

import (
	"bufio"
	"bytes"
	"testing"
)

func TestPrintfLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(bufio.NewWriter(&buf))
	err := w.PrintfLine("foo %d", 123)
	if s := buf.String(); s != "foo 123\n" || err != nil {
		t.Fatalf("s=%q; err=%s", s, err)
	}
}
