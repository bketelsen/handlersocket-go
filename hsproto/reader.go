// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hsproto

import (
	"bufio"
//	"bytes"
	"container/vector"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// BUG(rsc): To let callers manage exposure to denial of service
// attacks, Reader should allow them to set and reset a limit on
// the number of bytes read from the connection.

// A Reader implements convenience methods for reading requests
// or responses from a text protocol network connection.
type Reader struct {
	R   *bufio.Reader
	dot *dotReader
}

// NewReader returns a new Reader reading from r.
func NewReader(r *bufio.Reader) *Reader {
	return &Reader{R: r}
}

// ReadLine reads a single line from r,
// eliding the final \n or \r\n from the returned string.
func (r *Reader) ReadLine() (string, os.Error) {
	line, err := r.ReadLineBytes()
	return string(line), err
}

// ReadLineBytes is like ReadLine but returns a []byte instead of a string.
func (r *Reader) ReadLineBytes() ([]byte, os.Error) {
	r.closeDot()
	line, err := r.R.ReadBytes('\n')
	n := len(line)
	if n > 0 && line[n-1] == '\n' {
		n--
		if n > 0 && line[n-1] == '\r' {
			n--
		}
	}
	return line[0:n], err
}

// ReadContinuedLine reads a possibly continued line from r,
// eliding the final trailing ASCII white space.
// Lines after the first are considered continuations if they
// begin with a space or tab character.  In the returned data,
// continuation lines are separated from the previous line
// only by a single space: the newline and leading white space
// are removed.
//
// For example, consider this input:
//
//	Line 1
//	  continued...
//	Line 2
//
// The first call to ReadContinuedLine will return "Line 1 continued..."
// and the second will return "Line 2".
//
// A line consisting of only white space is never continued.
//
func (r *Reader) ReadContinuedLine() (string, os.Error) {
	line, err := r.ReadContinuedLineBytes()
	return string(line), err
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	n := len(s)
	for n > i && (s[n-1] == ' ' || s[n-1] == '\t') {
		n--
	}
	return s[i:n]
}

// ReadContinuedLineBytes is like ReadContinuedLine but
// returns a []byte instead of a string.
func (r *Reader) ReadContinuedLineBytes() ([]byte, os.Error) {
	// Read the first line.
	line, err := r.ReadLineBytes()
	if err != nil {
		return line, err
	}
	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}
	line = trim(line)

	// Look for a continuation line.
	c, err := r.R.ReadByte()
	if err != nil {
		// Delay err until we read the byte next time.
		return line, nil
	}
	if c != ' ' && c != '\t' {
		// Not a continuation.
		r.R.UnreadByte()
		return line, nil
	}

	// Read continuation lines.
	for {
		// Consume leading spaces; one already gone.
		for {
			c, err = r.R.ReadByte()
			if err != nil {
				break
			}
			if c != ' ' && c != '\t' {
				r.R.UnreadByte()
				break
			}
		}
		var cont []byte
		cont, err = r.ReadLineBytes()
		cont = trim(cont)
		line = append(line, ' ')
		line = append(line, cont...)
		if err != nil {
			break
		}

		// Check for leading space on next line.
		if c, err = r.R.ReadByte(); err != nil {
			break
		}
		if c != ' ' && c != '\t' {
			r.R.UnreadByte()
			break
		}
	}

	// Delay error until next call.
	if len(line) > 0 {
		err = nil
	}
	return line, err
}

func (r *Reader) readCodeLine(expectCode int) (code int, message string, err os.Error) {
	line, err := r.ReadLine()
	if err != nil {
		return
	}
	if len(line) < 3 || line[1] != ' ' {
		err = ProtocolError("short response: " + line)
		return
	}
	code, err = strconv.Atoi(line[0:1])
	if err != nil {
		err = ProtocolError("invalid response code: " + line)
		return
	}
	message = line[3:]
	if code != expectCode {
		err = &Error{code, message}
	}
	return
}

// ReadCodeLine reads a response code line of the form
//	code message
// where code is a 3-digit status code and the message
// extends to the rest of the line.  An example of such a line is:
//	220 plan9.bell-labs.com ESMTP
//
// If the prefix of the status does not match the digits in expectCode,
// ReadCodeLine returns with err set to &Error{code, message}.
// For example, if expectCode is 31, an error will be returned if
// the status is not in the range [310,319].
//
// If the response is multi-line, ReadCodeLine returns an error.
//
// An expectCode <= 0 disables the check of the status code.
//
func (r *Reader) ReadCodeLine(expectCode int) (code int, message string, err os.Error) {
	code, message, err = r.readCodeLine(expectCode)
	return
}

// ReadResponse reads a multi-line response of the form
//	code-message line 1
//	code-message line 2
//	...
//	code message line n
// where code is a 3-digit status code. Each line should have the same code.
// The response is terminated by a line that uses a space between the code and
// the message line rather than a dash. Each line in message is separated by
// a newline (\n).
//
// If the prefix of the status does not match the digits in expectCode,
// ReadResponse returns with err set to &Error{code, message}.
// For example, if expectCode is 31, an error will be returned if
// the status is not in the range [310,319].
//
// An expectCode <= 0 disables the check of the status code.
//
func (r *Reader) ReadResponse(expectCode int) (code int, message string, err os.Error) {
	code, message, err = r.readCodeLine(expectCode)
	for err == nil {
		var code2 int
		var moreMessage string
		code2, moreMessage, err = r.readCodeLine(expectCode)
		if code != code2 {
			err = ProtocolError("status code mismatch: " + strconv.Itoa(code) + ", " + strconv.Itoa(code2))
		}
		message += "\n" + moreMessage
	}
	return
}

// DotReader returns a new Reader that satisfies Reads using the
// decoded text of a dot-encoded block read from r.
// The returned Reader is only valid until the next call
// to a method on r.
//
// Dot encoding is a common framing used for data blocks
// in text protcols like SMTP.  The data consists of a sequence
// of lines, each of which ends in "\r\n".  The sequence itself
// ends at a line containing just a dot: ".\r\n".  Lines beginning
// with a dot are escaped with an additional dot to avoid
// looking like the end of the sequence.
//
// The decoded form returned by the Reader's Read method
// rewrites the "\r\n" line endings into the simpler "\n",
// removes leading dot escapes if present, and stops with error os.EOF
// after consuming (and discarding) the end-of-sequence line.
func (r *Reader) DotReader() io.Reader {
	r.closeDot()
	r.dot = &dotReader{r: r}
	return r.dot
}

type dotReader struct {
	r     *Reader
	state int
}

// Read satisfies reads by decoding dot-encoded data read from d.r.
func (d *dotReader) Read(b []byte) (n int, err os.Error) {

	br := d.r.R
	for n < len(b) {
		var c byte
		c, err = br.ReadByte()
		if err != nil {
			if err == os.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}
			if c == '\n' {
				continue
			}
		b[n] = c
		n++
	}

	if err != nil && d.r.dot == d {
		d.r.dot = nil
	}
	return
}

// closeDot drains the current DotReader if any,
// making sure that it reads until the ending dot line.
func (r *Reader) closeDot() {
	if r.dot == nil {
		return
	}
	buf := make([]byte, 128)
	for r.dot != nil {
		// When Read reaches EOF or an error,
		// it will set r.dot == nil.
		r.dot.Read(buf)
	}
}

// ReadDotBytes reads a dot-encoding and returns the decoded data.
//
// See the documentation for the DotReader method for details about dot-encoding.
func (r *Reader) ReadDotBytes() ([]byte, os.Error) {
	return ioutil.ReadAll(r.DotReader())
}

// ReadDotLines reads a dot-encoding and returns a slice
// containing the decoded lines, with the final \r\n or \n elided from each.
//
// See the documentation for the DotReader method for details about dot-encoding.
func (r *Reader) ReadDotLines() ([]string, os.Error) {
	// We could use ReadDotBytes and then Split it,
	// but reading a line at a time avoids needing a
	// large contiguous block of memory and is simpler.
	var v vector.StringVector
	var err os.Error
	for {
		var line string
		line, err = r.ReadLine()
		if err != nil {
			if err == os.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}

		// Dot by itself marks end; otherwise cut one dot.
		if len(line) > 0 && line[0] == '.' {
			if len(line) == 1 {
				break
			}
			line = line[1:]
		}
		v.Push(line)
	}
	return v, err
}





