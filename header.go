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
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"hz.tools/rf"
	"hz.tools/rfcap/internal"
	"hz.tools/sdr"
)

// MimeType is the rfcap v1 MIME type to be used.
const MimeType string = "application/x-hztools.rfcap"

// Magic signifies the rfcap Magic bytes. These are prefixed to the rfcap
// file, and can be used to determine if the file is valid rfcap or not.
type Magic [6]byte

var (
	// MagicVersion1 signifies the first version of rfcap.
	MagicVersion1 Magic = Magic{'R', 'F', 'C', 'A', 'P', '1'}
)

func (magic Magic) String() string {
	switch magic {
	case MagicVersion1:
		return "rfcap v1"
	default:
		return "unknown"
	}
}

// Size is the rfcap header size in Bytes.
var Size int = 48

// Header contains metadata around what the capture represents.
type Header struct {
	// Magic is 'RFCAP1'
	Magic Magic

	// CaptureTime signifies the time at which this capture was started.
	CaptureTime time.Time

	// CenterFrequency represents where the Center frequency of this capture
	// is centered.
	CenterFrequency rf.Hz

	// Number of Samples (each iq complex number is counted as a single sample)
	// per second.
	SampleRate uint

	// SampleFormat denotes what format this capture is in. It's useful to keep
	// iq information in its native capture format, and convert when required.
	SampleFormat sdr.SampleFormat

	// Compressed may only be set if the Sample Format is int16. If this is
	// true for any other Sample Format, this will result in an error being
	// returned.
	//
	// This assumes the data is actually 12 bits, and packs every 4th sample
	// into the other 3.
	Compressed bool

	// Endianness defines the ByteOrder used for the data in the rfcap
	// file.
	Endianness binary.ByteOrder
}

func (h Header) validate() error {
	if h.SampleFormat != sdr.SampleFormatU8 {
		if h.Endianness == nil {
			return fmt.Errorf("rfcap: rfcap.Header.Endianness must be set")
		}
	}

	if h.Compressed {
		if h.SampleFormat != sdr.SampleFormatI16 {
			return fmt.Errorf("rfcap: rfcap.Header.Compressed may only be set for i16")
		}
	}
	return nil
}

// HeaderFromSDR will create a Header from the provided SDR
func HeaderFromSDR(dev sdr.Sdr) (Header, error) {
	cf, err := dev.GetCenterFrequency()
	if err != nil {
		return Header{}, err
	}

	sps, err := dev.GetSampleRate()
	if err != nil {
		return Header{}, err
	}

	return Header{
		Magic:           MagicVersion1,
		CaptureTime:     time.Now(),
		CenterFrequency: cf,
		SampleRate:      sps,
		SampleFormat:    dev.SampleFormat(),
		Endianness:      internal.NativeEndian,
	}, nil
}

// Unmarshal will encode a header as Bytes.
func (h *Header) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)
	rh := rawHeader{}
	if err := binary.Read(buf, binary.LittleEndian, &rh); err != nil {
		return err
	}
	*h = rh.asExportHeader()
	return nil
}

// Marshal will encode a header as Bytes.
func (h *Header) Marshal() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := binary.Write(b, binary.LittleEndian, h.asBinaryHeader()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// rawHeader is the format that we actually i/o with. This lets us control
// the types we write out and be a bit more explicit about alignment. We always
// want to align to 128 bits in order to complex64 sample streams to maintain
// alignment if the consumer is not rfcap aware.
type rawHeader struct {
	Magic           [6]byte
	CaptureTime     int64
	CenterFrequency float64
	SampleRate      uint32
	SampleFormat    uint8
	Endianness      uint8
	Reserved        [20]uint8
}

func (h rawHeader) Validate() error {
	switch Magic(h.Magic) {
	case MagicVersion1:
		return nil
	default:
		return fmt.Errorf("Unknown rfcap version")
	}
}

// This will turn the regular Header into an rfcap "binary header" which is
// a bit more explicit when we do a binary.Write / binary.Read to and from a
// file.
func (h Header) asBinaryHeader() rawHeader {
	sampleFormat := uint8(h.SampleFormat)
	if h.Compressed {
		sampleFormat |= 128
	}

	return rawHeader{
		Magic:           [6]byte(h.Magic),
		CaptureTime:     h.CaptureTime.UnixNano(),
		CenterFrequency: float64(h.CenterFrequency),
		SampleRate:      uint32(h.SampleRate),
		SampleFormat:    sampleFormat,
		Endianness:      endianByteFromByteOrder(h.Endianness),
	}
}

const (
	byteOrderLittleEndian uint8 = 0
	byteOrderBigEndian    uint8 = 1
)

func endianByteFromByteOrder(bo binary.ByteOrder) uint8 {
	switch bo {
	case binary.LittleEndian:
		return byteOrderLittleEndian
	case binary.BigEndian:
		return byteOrderBigEndian
	default:
		// TODO(paultag): Add a warning
		return byteOrderLittleEndian
	}
}

func byteOrderFromEndianByte(bo uint8) binary.ByteOrder {
	switch bo {
	case byteOrderLittleEndian:
		return binary.LittleEndian
	case byteOrderBigEndian:
		return binary.BigEndian
	default:
		// TODO(paultag): Add a warning
		return binary.LittleEndian
	}
}

// This will translate the binary types into Go types that are much more sensible
// to work with (such as an rf.Hz or a time.Time)
func (h rawHeader) asExportHeader() Header {
	var (
		nanoseconds  int64 = 1e+9
		sampleFormat       = sdr.SampleFormat(h.SampleFormat)
		compressed   bool
	)

	if sampleFormat&128 == 128 {
		sampleFormat = sampleFormat ^ 128
		compressed = true
	}

	return Header{
		Magic:           Magic(h.Magic),
		CaptureTime:     time.Unix(h.CaptureTime/nanoseconds, h.CaptureTime%nanoseconds),
		CenterFrequency: rf.Hz(h.CenterFrequency),
		SampleRate:      uint(h.SampleRate),
		Compressed:      compressed,
		SampleFormat:    sampleFormat,
		Endianness:      byteOrderFromEndianByte(h.Endianness),
	}
}

// vim: foldmethod=marker
