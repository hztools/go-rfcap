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

package compress_test

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rfcap/internal/compress"
	"hz.tools/sdr"
)

func makeSine(in sdr.SamplesI16, sampleRate, freq float64) {
	for i := range in {
		ts := float64(i) / sampleRate
		sin, cos := math.Sincos(math.Pi * 2 * ts * freq)

		in[i] = [2]int16{
			int16(int32(cos*math.MaxInt16) & 0xFFF0),
			int16(int32(sin*math.MaxInt16) & 0xFFF0),
		}
	}
}

func TestCarrierWaveReader(t *testing.T) {
	var (
		in  = make(sdr.SamplesI16, 32*1024*48)
		out = make(sdr.SamplesI16, 32*1024*48)

		sampleRate float64 = 1000
		freq       float64 = 7
	)

	makeSine(in, sampleRate, freq)

	pipeReader, pipeWriter := sdr.Pipe(uint(sampleRate), sdr.SampleFormatI16)

	packedReader, err := compress.CompressReader(pipeReader)
	assert.NoError(t, err)

	plainReader, err := compress.DecompressReader(packedReader)
	assert.NoError(t, err)
	assert.Equal(t, uint(sampleRate), plainReader.SampleRate())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeWriter.Write(in)
	}()

	_, err = sdr.ReadFull(plainReader, out)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	wg.Wait()

	for i := range out {
		assert.Equal(t, in[i], out[i])
	}
}

func TestCarrierWaveWriter(t *testing.T) {
	var (
		in  = make(sdr.SamplesI16, 32*1024*48)
		out = make(sdr.SamplesI16, 32*1024*48)

		sampleRate float64 = 1000
		freq       float64 = 7
	)

	makeSine(in, sampleRate, freq)

	pipeReader, pipeWriter := sdr.Pipe(uint(sampleRate), sdr.SampleFormatI16)

	plainReader, err := compress.DecompressReader(pipeReader)
	assert.NoError(t, err)
	assert.Equal(t, uint(sampleRate), plainReader.SampleRate())

	packedWriter, err := compress.CompressWriter(pipeWriter)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		packedWriter.Write(in)
	}()

	_, err = sdr.ReadFull(plainReader, out)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	wg.Wait()

	for i := range out {
		assert.Equal(t, in[i], out[i])
	}
}

// vim: foldmethod=marker
