// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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

package uring

// #include <liburing.h>
import "C"

import (
	"log"
	"os"
	"unsafe"

	"hz.tools/rfcap"
	"hz.tools/sdr"
	"hz.tools/sdr/yikes"
)

type Reader struct {
	ring   *Ring
	f      *os.File
	opts   ReaderOpts
	header rfcap.Header

	iqSize int
	buf    unsafe.Pointer
	iq     sdr.Samples
	offset int
	read   int

	seen func()
}

type ReaderOpts struct {
	IQLength int
}

func (r *Ring) Reader(f *os.File, opts ReaderOpts) (*Reader, error) {
	header, err := rfcap.ReadHeader(f)
	if err != nil {
		return nil, err
	}

	var (
		iqSize = header.SampleFormat.Size() * opts.IQLength
		buf    = C.malloc(C.size_t(iqSize))
		offset = rfcap.Size
	)
	iq, err := yikes.Samples(uintptr(buf), opts.IQLength, header.SampleFormat)
	if err != nil {
		return nil, err
	}

	rdr := &Reader{
		f:      f,
		ring:   r,
		header: header,

		iqSize: iqSize,
		offset: offset,
		buf:    buf,
		iq:     iq,
	}

	return rdr, nil
}

func (r *Reader) Next() bool {
	var (
		sqe *C.struct_io_uring_sqe = ioUringGetSQE(r.ring.ring)
		cqe *C.struct_io_uring_cqe
	)

	if r.seen != nil {
		r.seen()
		r.seen = nil
	}

	ioUringPrepRead(sqe, r.f.Fd(), r.buf, r.iqSize, r.offset)
	ioUringSubmit(r.ring.ring)

	if err := ioUringWaitCQE(r.ring.ring, &cqe); err != nil {
		log.Printf("io_uring_wait_cpe failed: %s", err)
		log.Printf(" offset: %d", r.offset)
		return false
	}
	n := int(cqe.res)

	if n == 0 {
		return false
	}

	if n%r.iq.Format().Size() != 0 {
		log.Printf("read isn't aligned to iq bounds, abort")
		return false
	}

	r.seen = func() {
		ioUringCQESeen(r.ring.ring, cqe)
	}

	r.read = n
	r.offset += n
	return true
}

func (r *Reader) Samples() sdr.Samples {
	return r.iq.Slice(0, r.read/r.iq.Format().Size())
}

func (r *Reader) Close() error {
	// free
	return nil
}

// vim: foldmethod=marker
