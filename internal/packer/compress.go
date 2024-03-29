// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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

package packer

import (
	"fmt"
	"unsafe"

	"hz.tools/sdr"
)

func unsafeReallyDontDoThis(iq sdr.SamplesI16) []int16 {
	base := uintptr(unsafe.Pointer(&iq[0]))
	size := iq.Length() * 2
	var b = struct {
		addr uintptr
		len  int
		cap  int
	}{base, size, size}
	return *(*[]int16)(unsafe.Pointer(&b))
}

// DecompressI16 will take our sdr.SamplesI16 type (under the hood it's a
// [2]int16 since it contains both the real and imaginary components)
// and perform a decompression. Every 4th IQ value (which is to say, the
// last imaginary part) is smeared across the LSB of the other values
// (assuming this is an int12). This will unpack that last component and
// re-create well-formed IQ data.
func DecompressI16(in, out sdr.SamplesI16) (int, error) {
	n, err := Decompress(
		unsafeReallyDontDoThis(in),
		unsafeReallyDontDoThis(out),
	)
	if n%2 != 0 {
		return n / 2, fmt.Errorf("decompress: target iq length is misaligned")
	}
	return n / 2, err
}

// CompressI16 will take our sdr.SamplesI16 type, and take every other
// imaginary value, and smear the 12 bits of goodness across the other 3
// values. This is then repackaged and sent out as if everything is totally
// normal, totally fine.
func CompressI16(in, out sdr.SamplesI16) (int, error) {
	n, err := Compress(
		unsafeReallyDontDoThis(in),
		unsafeReallyDontDoThis(out),
	)
	if n%2 != 0 {
		return n / 2, fmt.Errorf("compress: target iq length is misaligned")
	}

	return n / 2, err
}

// Compress will write the IQ samples in the "iniq" buffer, and smear
// every fourth int16 (really int12) into the other 3 int12 values. The
// output will be written to the output buffer.
func Compress(iniq, outpacked []int16) (int, error) {
	if len(iniq)%4 != 0 {
		return 0, fmt.Errorf("compress: input iq length isn't a multiple of 4")
	}

	var (
		in     = *(*[]uint16)(unsafe.Pointer(&iniq))
		outlen = (len(in) / 4) * 3
		out    = *(*[]uint16)(unsafe.Pointer(&outpacked))
		buf    = make([]uint16, 3)
	)

	if len(out) < outlen {
		return 0, fmt.Errorf("compress: output buffer isn't large enough")
	}

	for i := 0; i < len(in)/4; i++ {
		inI := i * 4
		outI := i * 3

		buf[0] = (in[inI+0] & 0xFFF0) | ((in[inI+3] & 0x00F0) >> 4)
		buf[1] = (in[inI+1] & 0xFFF0) | ((in[inI+3] & 0x0F00) >> 8)
		buf[2] = (in[inI+2] & 0xFFF0) | ((in[inI+3] & 0xF000) >> 12)

		copy(out[outI:], buf)
	}

	return outlen, nil
}

// Decompress will read the input packed int12 values (packed as per
// the 'Compress' routine in this same package) and unpack the 4th
// int16 (really int12) value as its own phasor.
func Decompress(inpacked, outiq []int16) (int, error) {
	if len(inpacked)%3 != 0 {
		return 0, fmt.Errorf("decompress: input compressed length isn't a multiple of 3")
	}

	var (
		in     = *(*[]uint16)(unsafe.Pointer(&inpacked))
		outlen = (len(in) / 3) * 4
		out    = *(*[]uint16)(unsafe.Pointer(&outiq))
		buf    = make([]uint16, 4)
	)

	if len(out) < outlen {
		return 0, fmt.Errorf("compress: output buffer isn't large enough")
	}

	for i := 0; i < len(in)/3; i++ {
		inI := i * 3
		outI := i * 4

		buf[0] = (in[inI+0] & 0xFFF0)
		buf[1] = (in[inI+1] & 0xFFF0)
		buf[2] = (in[inI+2] & 0xFFF0)
		buf[3] = (in[inI+0]&0x000F)<<4 |
			(in[inI+1]&0x000F)<<8 |
			(in[inI+2]&0x000F)<<12

		copy(out[outI:], buf)
	}

	return outlen, nil
}

// vim: foldmethod=marker
