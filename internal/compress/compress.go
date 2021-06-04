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

package internal

import (
	"unsafe"
)

//
func Compress(iniq []int16) []int16 {
	if len(iniq)%4 != 0 {
		panic("Not aligned")
	}

	var (
		in  []uint16 = *(*[]uint16)(unsafe.Pointer(&iniq))
		out          = make([]uint16, (len(in)/4)*3)
		buf          = make([]uint16, 3)
	)

	for i := 0; i < len(in)/4; i++ {
		inI := i * 4
		outI := i * 3

		buf[0] = (in[inI+0] & 0xFFF0) | ((in[inI+3] & 0x00F0) >> 4)
		buf[1] = (in[inI+1] & 0xFFF0) | ((in[inI+3] & 0x0F00) >> 8)
		buf[2] = (in[inI+2] & 0xFFF0) | ((in[inI+3] & 0xF000) >> 12)

		copy(out[outI:], buf)
	}

	var outiq []int16 = *(*[]int16)(unsafe.Pointer(&out))
	return outiq
}

//
func Decompress(iniq []int16) []int16 {
	if len(iniq)%3 != 0 {
		panic("Not aligned")
	}

	var (
		in  []uint16 = *(*[]uint16)(unsafe.Pointer(&iniq))
		out          = make([]uint16, (len(in)/3)*4)
		buf          = make([]uint16, 4)
	)

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

	var outiq []int16 = *(*[]int16)(unsafe.Pointer(&out))
	return outiq
}

// vim: foldmethod=marker
