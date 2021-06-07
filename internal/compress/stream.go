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

package compress

import (
	"fmt"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func CompressReader(in sdr.Reader) (sdr.Reader, error) {
	if in.SampleFormat() != sdr.SampleFormatI16 {
		return nil, fmt.Errorf("compress: only i16 supported")
	}

	return stream.ReadTransformer(in, stream.ReadTransformerConfig{
		InputBufferLength:  32 * 1024,
		OutputBufferLength: ((32 * 1024) / 4) * 3,
		OutputSampleRate:   in.SampleRate(),
		OutputSampleFormat: in.SampleFormat(),
		Proc: func(in, out sdr.Samples) (int, error) {
			return CompressI16(in.(sdr.SamplesI16), out.(sdr.SamplesI16))
		},
	})
}

func DecompressReader(in sdr.Reader) (sdr.Reader, error) {
	if in.SampleFormat() != sdr.SampleFormatI16 {
		return nil, fmt.Errorf("compress: only i16 supported")
	}

	return stream.ReadTransformer(in, stream.ReadTransformerConfig{
		InputBufferLength:  ((32 * 1024) / 4) * 3,
		OutputBufferLength: 32 * 1024,
		OutputSampleRate:   in.SampleRate(),
		OutputSampleFormat: in.SampleFormat(),
		Proc: func(in, out sdr.Samples) (int, error) {
			return DecompressI16(in.(sdr.SamplesI16), out.(sdr.SamplesI16))
		},
	})
}

// vim: foldmethod=marker