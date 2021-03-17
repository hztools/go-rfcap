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
	"io"

	"github.com/pierrec/lz4/v4"
)

// Compression is used to indicate the type of compression used in an rfcap
// stream. Compression is expensive to write, but eases storage requirements.
type Compression uint16

func (c Compression) String() string {
	switch c {
	case NoCompression:
		return "none"
	case Lz4Compression:
		return "lz4"
	default:
		return "unknown"
	}
}

// Reader will take an io.Reader, and decompress that stream, providing an
// io.Reader of the uncompressed plaintext.
func (c Compression) Reader(r io.Reader) io.Reader {
	switch c {
	case Lz4Compression:
		zr := lz4.NewReader(r)
		zr.Reset(r)
		zr.Apply(lz4.ConcurrencyOption(-1))
		return zr
	// case NoCompression is default:
	default:
		return r
	}
}

// Writer will take an io.Writer, and compress any writes to that stream.
func (c Compression) Writer(w io.Writer) io.Writer {
	switch c {
	case Lz4Compression:
		zw := lz4.NewWriter(w)
		zw.Apply(
			lz4.CompressionLevelOption(lz4.Level6),
			lz4.ConcurrencyOption(-1),
		)
		return zw

	// case NoCompression is default:
	default:
		return w
	}
}

var (
	// NoCompression means the IQ samples are written to disk raw, without
	// any particular
	NoCompression Compression = 0

	// Lz4Compression will use the LZ4 algorithm to compress or decompress
	// the rfcap file.
	Lz4Compression Compression = 1
)
