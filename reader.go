// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
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

	"hz.tools/sdr"
)

// reader is the internal type that implements the sdr.Reader interface. This
// will force the caller to handle Closing, etc.
type reader struct {
	header Header
	r      sdr.Reader
}

// Reader will create a new sdr.Reader from the provided io stream.
func Reader(in io.Reader) (sdr.Reader, Header, error) {
	header := rawHeader{}

	// TODO(paultag): this ought to be big endian maybe? Network order and
	// all that.
	if err := binary.Read(in, binary.LittleEndian, &header); err != nil {
		return nil, Header{}, err
	}

	if err := header.Validate(); err != nil {
		return nil, Header{}, err
	}

	h := header.asExportHeader()

	return reader{
		header: h,
		r:      sdr.ByteReader(in, h.Endianness, h.SampleRate, h.SampleFormat),
	}, h, nil
}

func (r reader) SampleRate() uint {
	return r.header.SampleRate
}

// SampleFormat will return the SampleFormat of the underlying stream.
func (r reader) SampleFormat() sdr.SampleFormat {
	return r.header.SampleFormat
}

// Read implements the sdr.Reader interface.
func (r reader) Read(samples sdr.Samples) (int, error) {
	return r.r.Read(samples)
}

// vim: foldmethod=marker
