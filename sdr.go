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

	"hz.tools/rf"
	"hz.tools/sdr"
)

// ReaderSdr will return a fake "SDR" that complies with the sdr.Sdr interface,
// where StartRx will provide the rfcap Reader. There are a number of read
// only attributes (frequency, samples per second), and calls to a number of
// methods will return sdr.ErrNotSupported.
func ReaderSdr(in io.Reader) (sdr.Receiver, error) {
	reader, header, err := Reader(in)
	if err != nil {
		return nil, err
	}

	return fakeSdr{
		header: header,
		reader: reader,
	}, nil
}

type fakeSdr struct {
	header Header
	reader sdr.Reader
}

func (s fakeSdr) HardwareInfo() sdr.HardwareInfo {
	return sdr.HardwareInfo{}
}

func (s fakeSdr) Close() error {
	return nil
}

func (s fakeSdr) GetCenterFrequency() (rf.Hz, error) {
	return s.header.CenterFrequency, nil
}

func (s fakeSdr) GetSampleRate() (uint32, error) {
	return s.header.SampleRate, nil
}

func (s fakeSdr) SampleFormat() sdr.SampleFormat {
	return s.header.SampleFormat
}

type nopCloser struct {
	sdr.Reader
}

func (nopCloser) Close() error { return nil }

func newNopCloser(r sdr.Reader) sdr.ReadCloser {
	return nopCloser{Reader: r}
}

func (s fakeSdr) StartRx() (sdr.ReadCloser, error) {
	return newNopCloser(s.reader), nil
}

func (s fakeSdr) SetCenterFrequency(rf.Hz) error         { return sdr.ErrNotSupported }
func (s fakeSdr) SetAutomaticGain(bool) error            { return sdr.ErrNotSupported }
func (s fakeSdr) GetGainStages() (sdr.GainStages, error) { return nil, nil }
func (s fakeSdr) GetGain(sdr.GainStage) (float32, error) { return 0, sdr.ErrNotSupported }
func (s fakeSdr) SetGain(sdr.GainStage, float32) error   { return sdr.ErrNotSupported }
func (s fakeSdr) SetSampleRate(uint32) error             { return sdr.ErrNotSupported }
func (s fakeSdr) SetPPM(int) error                       { return sdr.ErrNotSupported }

// vim: foldmethod=marker
