package slog

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var seed int64

func TestMain(m *testing.M) {
	seed = time.Now().UnixNano()
	flag.Int64Var(&seed, "seed", seed, "random seed")
	flag.Parse()
	rand.Seed(seed)
	fmt.Printf("using random seed %d (0x%x)\n", seed, seed)
	res := m.Run()
	os.Exit(res)
}

// returns the number of times push & pop failed
func popWaitPush(bp *bpool, n int, w time.Duration) (int, int) {
	var popf, pushf int
	for i := 0; i < n; i++ {
		sb := bp.pop()
		//os.Stdout.Sync()
		if sb == nil {
			popf++
			// alloc a new one by hand
			sb = &bytes.Buffer{}
			sb.Grow(BufSz)
		}
		if w > 0 {
			time.Sleep(w)
		}
		if !bp.push(sb) {
			pushf++
		}
	}
	return pushf, popf
}

func TestBPoolParallel(t *testing.T) {
	var bp bpool
	var pushFailed, popFailed uint64

	if !bp.init() {
		t.Fatalf("bpool init failed\n")
	}
	initFree := bp.free()
	if initFree != BufPoolSz {
		t.Fatalf("bpool unexpected initial free buffers: %d, expected %d\n",
			initFree, BufPoolSz)
	}

	//threads := 100
	//loops := 100000
	threads := rand.Intn(256) + 1
	loops := rand.Intn(100000) + 1

	t.Logf("using %d threads and %d loops, seed %d\n", threads, loops, seed)
	start := make(chan struct{})
	waitgrp := &sync.WaitGroup{}
	for i := 0; i < threads; i++ {
		waitgrp.Add(1)
		go func() {
			defer waitgrp.Done()

			<-start // wait for start
			pushf, popf := popWaitPush(&bp, loops, 0)
			atomic.AddUint64(&pushFailed, uint64(pushf))
			atomic.AddUint64(&popFailed, uint64(popf))
		}()
	}
	close(start)
	waitgrp.Wait()
	os.Stdout.Sync()
	if pushFailed != popFailed {
		t.Errorf("no of push failed != pop failed: %d != %d, seed %d\n",
			pushFailed, popFailed, seed)
	}
	if bp.free() != initFree {
		t.Errorf("final free buffers != initial valued: %d != %d, seed %d\n",
			bp.free(), initFree, seed)
	}
}
