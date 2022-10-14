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
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	"hz.tools/rfcap"
)

type Reader struct {
	ring   *Ring
	f      *os.File
	opts   ReaderOpts
	header rfcap.Header

	blocks int
	iqSize int
	buf    unsafe.Pointer
	iovecs unsafe.Pointer
	offset int
}

type ReaderOpts struct {
	BlockSize int
	IQLength  int
}

func (r *Ring) Reader(f *os.File, opts ReaderOpts) (*Reader, error) {
	header, err := rfcap.ReadHeader(f)
	if err != nil {
		return nil, err
	}

	var (
		iqSize      = header.SampleFormat.Size() * opts.IQLength
		iqBlockSize = opts.BlockSize
		// iqBlockLen  = opts.BlockSize / header.SampleFormat.Size()
		blocks = iqSize / opts.BlockSize
		buf    = C.malloc(C.size_t(iqSize))
		iovecs = C.malloc(C.size_t(uintptr(blocks) * unsafe.Sizeof(C.struct_iovec{})))
		offset = rfcap.Size
	)

	if iqSize%opts.BlockSize != 0 {
		return nil, fmt.Errorf("Reader: BlockSize is misaligned to IQ Format")
	}

	rdr := &Reader{
		f:      f,
		ring:   r,
		header: header,

		blocks: blocks,
		iqSize: iqSize,
		offset: offset,
		buf:    buf,
		iovecs: iovecs,
	}

	iovs := rdr.getIovecs()
	for i := range iovs {
		iovs[i].iov_base = unsafe.Pointer(uintptr(buf) + uintptr(i*iqBlockSize))
		iovs[i].iov_len = C.size_t(iqBlockSize)
	}

	return rdr, nil
}

func (r *Reader) getIovecs() []C.struct_iovec {
	var b = struct {
		base uintptr
		len  int
		cap  int
	}{uintptr(r.iovecs), r.blocks, r.blocks}
	iovecs := *(*[]C.struct_iovec)(unsafe.Pointer(&b))
	return iovecs
}

func (r *Reader) Next() bool {
	var (
		sqe    *C.struct_io_uring_sqe = C.io_uring_get_sqe(r.ring.ring)
		iovecs                        = r.getIovecs()
	)

	C.io_uring_prep_readv(
		sqe,
		C.int(r.f.Fd()),
		&iovecs[0],
		C.uint(r.blocks),
		C.ulonglong(r.offset),
	)
	C.io_uring_submit(r.ring.ring)

	var cqe *C.struct_io_uring_cqe
	if errno := C.io_uring_wait_cqe(r.ring.ring, &cqe); errno < 0 {
		log.Printf("io_uring_wait_cpe failed: %s", syscall.Errno(-errno))
		log.Printf(" offset: %d", r.offset)
		return false
	}

	// for _, iov := range iovecs {
	// 	log.Printf("%d", iov.iov_len)
	// }

	r.offset += r.iqSize
	C.io_uring_cqe_seen(r.ring.ring, cqe)
	return true
}

func (r *Reader) Close() error {
	// free
	return nil
}

// vim: foldmethod=marker
