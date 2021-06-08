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

package packer_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rfcap/internal/packer"
	"hz.tools/sdr"
)

func TestDecompressBadLength(t *testing.T) {
	v := make([]int16, 5)
	o := make([]int16, 10)
	_, err := packer.Decompress(v, o)
	assert.Error(t, err)
}

func TestDecompressShortOutput(t *testing.T) {
	v := make([]int16, 4)
	o := make([]int16, 2)
	_, err := packer.Decompress(v, o)
	assert.Error(t, err)
}

func TestCompressBadLength(t *testing.T) {
	v := make([]int16, 5)
	o := make([]int16, 10)
	_, err := packer.Compress(v, o)
	assert.Error(t, err)
}

func TestCompressShortOutput(t *testing.T) {
	v := make([]int16, 4)
	o := make([]int16, 2)
	_, err := packer.Compress(v, o)
	assert.Error(t, err)
}

func TestCompressAll(t *testing.T) {
	var (
		i  int16
		pv = make([]int16, 3)
		ov = make([]int16, 4)
	)

	for i = math.MinInt16; i < math.MaxInt16; i++ {
		i := i & -16

		v := []int16{i, i, i, i}

		n, err := packer.Compress(v, pv)
		assert.NoError(t, err)
		assert.Equal(t, 3, n)
		n, err = packer.Decompress(pv, ov)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, v, ov)
	}
}

func TestCompressIQAll(t *testing.T) {
	in := make(sdr.SamplesI16, 32*1024)
	packed := make(sdr.SamplesI16, ((32*1024)/4)*3)
	out := make(sdr.SamplesI16, 32*1024)

	for i := range in {
		in[i] = [2]int16{
			0x0AB0,
			0x0AB0,
		}
	}

	n, err := packer.CompressI16(in, packed)
	assert.NoError(t, err)
	assert.Equal(t, len(packed), n)
	n, err = packer.DecompressI16(packed, out)
	assert.NoError(t, err)
	assert.Equal(t, len(out), n)

	for i := range out {
		assert.Equal(t, [2]int16{
			0x0AB0,
			0x0AB0,
		}, out[i])
	}
}

// vim: foldmethod=marker
