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

// #cgo pkg-config: liburing
//
// #include <liburing.h>
import "C"

import (
	"syscall"
	"unsafe"
)

func rvToErr(rv C.int) error {
	if rv < 0 {
		return syscall.Errno(-rv)
	}
	return nil
}

func ioUringQueueExit(ring *C.struct_io_uring) error {
	C.io_uring_queue_exit(ring)
	return nil
}

func ioUringQueueInit(depth uint, ring *C.struct_io_uring, flags uint) error {
	return rvToErr(C.io_uring_queue_init(C.uint(depth), ring, C.uint(flags)))
}

func ioUringGetSQE(ring *C.struct_io_uring) *C.struct_io_uring_sqe {
	return C.io_uring_get_sqe(ring)
}

func ioUringPrepRead(
	sqe *C.struct_io_uring_sqe,
	fd uintptr,
	buf unsafe.Pointer,
	bufSize int,
	offset int,
) error {
	C.io_uring_prep_read(
		sqe,
		C.int(fd),
		buf,
		C.uint(bufSize),
		C.ulonglong(offset),
	)
	return nil
}

func ioUringSubmit(ring *C.struct_io_uring) error {
	return rvToErr(C.io_uring_submit(ring))
}

func ioUringWaitCQE(ring *C.struct_io_uring, cpe **C.struct_io_uring_cqe) error {
	for {
		if err := rvToErr(C.io_uring_wait_cqe(ring, cpe)); err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		return nil
	}
}

func ioUringCQESeen(ring *C.struct_io_uring, cpe *C.struct_io_uring_cqe) error {
	C.io_uring_cqe_seen(ring, cpe)
	return nil
}

// vim: foldmethod=marker
