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
)

type Ring struct {
	ring *C.struct_io_uring
}

func (r *Ring) Close() error {
	C.io_uring_queue_exit(r.ring)
	return nil
}

func NewRing() (*Ring, error) {
	var (
		ring       C.struct_io_uring
		queueDepth int = 32
	)

	if errno := C.io_uring_queue_init(C.uint(queueDepth), &ring, 0); errno != 0 {
		return nil, syscall.Errno(-errno)
	}

	return &Ring{
		ring: &ring,
	}, nil
}

// vim: foldmethod=marker
