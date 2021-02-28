// {{{ Copyright (c) Paul R. Tagliamonte <paul@kc3nwj.com>, 2020
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE. }}}

package rfcap

import (
	"encoding/binary"
	"io"

	"kc3nwj.com/sdr"
)

type writer struct {
	header Header
	w      sdr.Writer
}

// Writer will create a new sdr.Reader that writes to the underlying Stream.
func Writer(out io.Writer, header Header) (sdr.Writer, error) {
	if err := header.validate(); err != nil {
		return nil, err
	}

	bh := header.asBinaryHeader()
	if err := bh.Validate(); err != nil {
		return nil, err
	}
	if err := binary.Write(out, binary.LittleEndian, bh); err != nil {
		return nil, err
	}

	return writer{
		header: header,
		w:      sdr.ByteWriter(out, header.Endianness, header.SampleRate, header.SampleFormat),
	}, nil
}

func (w writer) SampleRate() uint32 {
	return w.header.SampleRate
}

// SampleFormat will return the sample format being encoded in this stream.
func (w writer) SampleFormat() sdr.SampleFormat {
	return w.header.SampleFormat
}

// Write implements the sdr.Writer format.
func (w writer) Write(samples sdr.Samples) (int, error) {
	return w.w.Write(samples)
}

// vim: foldmethod=marker
