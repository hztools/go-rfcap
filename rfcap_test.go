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

package rfcap_test

import (
	"context"
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/rfcap"
	"hz.tools/sdr"
	"hz.tools/sdr/mock"
)

func TestHeaderFromSdr(t *testing.T) {
	mockSdr := mock.New(mock.Config{})

	assert.NoError(t, mockSdr.SetSampleRate(10e6))
	assert.NoError(t, mockSdr.SetCenterFrequency(1090*rf.MHz))

	header, err := rfcap.HeaderFromSDR(mockSdr)
	assert.NoError(t, err)

	assert.Equal(t, header.CenterFrequency, 1090*rf.MHz)
	assert.Equal(t, header.SampleRate, uint32(10e6))
}

func TestRfcapEndianO(t *testing.T) {
	fd, err := ioutil.TempFile("", "go-rf-rfcap_test")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer os.Remove(fd.Name())

	when := time.Now()

	writer, err := rfcap.Writer(fd, rfcap.Header{
		Magic:           rfcap.MagicVersion1,
		CaptureTime:     when,
		CenterFrequency: rf.MustParseHz("1337MHz"),
		SampleRate:      1.8e+8,
		SampleFormat:    sdr.SampleFormatC64,
		Endianness:      binary.LittleEndian,
	})
	assert.NoError(t, err)

	refSamples := sdr.SamplesC64{
		complex(0.0, 0.0i),
	}

	n, err := writer.Write(refSamples)
	assert.NoError(t, err)
	assert.Equal(t, len(refSamples), n)
}

func TestRfcapHeaderIO(t *testing.T) {
	fd, err := ioutil.TempFile("", "go-rf-rfcap_test")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer os.Remove(fd.Name())

	when := time.Now()

	writer, err := rfcap.Writer(fd, rfcap.Header{
		Magic:           rfcap.MagicVersion1,
		CaptureTime:     when,
		CenterFrequency: rf.MustParseHz("1337MHz"),
		SampleRate:      1.8e+8,
		SampleFormat:    sdr.SampleFormatU8,
	})
	assert.NoError(t, err)

	refSamples := sdr.SamplesU8{
		[2]uint8{1, 2},
		[2]uint8{3, 4},
		[2]uint8{4, 3},
		[2]uint8{2, 1},
	}

	n, err := writer.Write(refSamples)
	assert.NoError(t, err)
	assert.Equal(t, len(refSamples), n)

	_, err = fd.Seek(0, 0)
	assert.NoError(t, err)

	reader, header, err := rfcap.Reader(fd)
	assert.NoError(t, err)

	assert.True(t, header.CaptureTime.Equal(when))
	assert.Equal(t, header.CenterFrequency, rf.Hz(1337e+6))
	assert.Equal(t, header.SampleRate, uint32(1.8e+8))

	outSamples := make(sdr.SamplesU8, len(refSamples))
	n, err = reader.Read(outSamples)
	assert.NoError(t, err)
	assert.Equal(t, len(refSamples), n)

	assert.Equal(t, outSamples, refSamples)
}

func TestRfcapSDR(t *testing.T) {
	fd, err := ioutil.TempFile("", "go-rf-rfcap_test")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer os.Remove(fd.Name())

	when := time.Now()

	writer, err := rfcap.Writer(fd, rfcap.Header{
		Magic:           rfcap.MagicVersion1,
		CaptureTime:     when,
		CenterFrequency: rf.MustParseHz("1337MHz"),
		SampleRate:      1.8e+8,
		SampleFormat:    sdr.SampleFormatU8,
	})
	assert.NoError(t, err)
	refSamples := sdr.SamplesU8{
		[2]uint8{1, 2},
		[2]uint8{3, 4},
		[2]uint8{4, 3},
		[2]uint8{2, 1},
	}

	n, err := writer.Write(refSamples)
	assert.NoError(t, err)
	assert.Equal(t, len(refSamples), n)

	_, err = fd.Seek(0, 0)
	assert.NoError(t, err)

	fakeSdr, err := rfcap.ReaderSdr(fd)
	assert.NoError(t, err)

	centerFreq, err := fakeSdr.GetCenterFrequency()
	assert.NoError(t, err)
	samplesPerSecond, err := fakeSdr.GetSampleRate()
	assert.NoError(t, err)

	assert.Equal(t, centerFreq, rf.Hz(1337e+6))
	assert.Equal(t, samplesPerSecond, uint32(1.8e+8))

	rx, err := fakeSdr.StartRx(context.TODO())
	assert.NoError(t, err)

	outSamples := make(sdr.SamplesU8, len(refSamples))
	n, err = sdr.ReadFull(rx, outSamples)
	assert.NoError(t, err)
	assert.Equal(t, len(refSamples), n)
	assert.Equal(t, outSamples, refSamples)
}

// vim: foldmethod=marker
