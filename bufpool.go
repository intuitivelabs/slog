// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

package slog

import (
	"bytes"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// keep a pool of bytes.Buffer(s) for fast re-use
const (
	BufPoolSz = 128 // maximum buffer in pool, must be 2^k
	BufSz     = 1024
	MaxBufSz  = 65534

//	BufSz     = 1024  // typical buf sz
//	MaxBufSz  = 65534 // max allowed (if more reset)
)

type bpool struct {
	// BufPoolSz must be 2^k to avoid wraparounds problem with first & last
	// and module BufPoolSz.
	pool [BufPoolSz]*bytes.Buffer
	// always pop from tail%BufPoolSz
	tail uint64 // first free (idx is module BufPoolSZ) (always increases)
	// always push at head%BufPoolSz
	head uint64 // last free (always increases)
}

func (p *bpool) idx(h uint64) uint64 {
	return h % BufPoolSz
}

func (p *bpool) push(sb *bytes.Buffer) bool {
	if sb == nil {
		return false
	}
	var h uint64
	for {
		t := atomic.LoadUint64(&p.tail)
		h = atomic.LoadUint64(&p.head)
		if (h - t) >= BufPoolSz {
			return false // full
		}
		// increase
		if atomic.CompareAndSwapUint64(&p.head, h, h+1) {
			// got  a safe slot for storing
			break
		}
	}
	i := p.idx(h)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&p.pool[i])),
		unsafe.Pointer(sb))
	return true
}

func (p *bpool) pop() *bytes.Buffer {
	var ret unsafe.Pointer
	var n uint
	for {
		h := atomic.LoadUint64(&p.head)
		t := atomic.LoadUint64(&p.tail)
		if h == t {
			// empty
			return nil
		}
		// try to take ownership of the pointer
		i := p.idx(t)
		ret = atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&p.pool[i])),
			nil)
		if ret != nil {
			atomic.AddUint64(&p.tail, 1)
			break
		}
		// else some other pop() won this race => retry
		n++
		runtime.Gosched()
	}
	return (*bytes.Buffer)(ret)
}

func (p *bpool) free() uint64 {
	h := atomic.LoadUint64(&p.head)
	t := atomic.LoadUint64(&p.tail)
	return h - t
}

func (p *bpool) init() bool {
	if p.tail != 0 || p.head != 0 {
		return false
	}
	for i := 0; i < BufPoolSz; i++ {
		buf := make([]byte, 0, BufSz)
		sb := bytes.NewBuffer(buf)
		if !p.push(sb) {
			return false
		}
	}
	return true
}
