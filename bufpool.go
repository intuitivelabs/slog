// Copyright 2021 Intuitive Labs GmbH. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE.txt file in the root of the source
// tree.

package slog

import (
	"bytes"
	"fmt"
	"sync"
)

// keep a pool of bytes.Buffer(s) for fast re-use
const (
	BufPoolSz = 128   // maximum buffer in pool, must be 2^k
	BufSz     = 1024  // typical buf sz
	MaxBufSz  = 65534 // max allowed (if more reset)
)

type bpool struct {
	lock sync.Mutex
	// always push at head%BufPoolSz
	head uint64 // last free (always increases)
	// always pop from tail%BufPoolSz
	tail uint64 // first free (idx is module BufPoolSZ) (always increases)
	// BufPoolSz must be 2^k to avoid wraparounds problem with first & last
	// and module BufPoolSz.
	pool [BufPoolSz]*bytes.Buffer
}

func (p *bpool) idx(h uint64) uint64 {
	return h % BufPoolSz
}

func (p *bpool) push(sb *bytes.Buffer) bool {
	if sb == nil {
		return false
	}
	p.lock.Lock()
	t := p.tail
	h := p.head
	if (h - t) >= BufPoolSz {
		p.lock.Unlock()
		return false // full
	}
	// increase
	p.head++
	i := p.idx(h)
	// sanity checks
	if p.pool[i] != nil {
		errStr := fmt.Sprintf("push(): unexpected non-null pointer at h: %d"+
			" (crt head %d), tail %d (crt %d) i %d : %p\n",
			h, p.head, t, p.tail, i, p.pool[i])
		p.lock.Unlock()
		panic(errStr)
		return false
	}
	p.pool[i] = sb
	p.lock.Unlock()
	return true
}

func (p *bpool) pop() *bytes.Buffer {
	p.lock.Lock()
	h := p.head
	t := p.tail
	if h == t {
		// empty
		p.lock.Unlock()
		return nil
	}
	// "reserve" a tail index
	p.tail++
	i := p.idx(t)
	ret := p.pool[i]
	// sanity checks
	if ret == nil {
		errStr := fmt.Sprintf("pop(): unexpected null pointer at h: %d"+
			" (crt head %d), tail %d (crt %d) i %d : %p\n",
			h, p.head, t, p.tail, i, p.pool[i])
		p.lock.Unlock()
		panic(errStr)
		return nil
	}
	p.pool[i] = nil
	p.lock.Unlock()
	return ret
}

func (p *bpool) free() uint64 {
	p.lock.Lock()
	ret := p.head - p.tail
	p.lock.Unlock()
	return ret
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
