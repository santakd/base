// Code generated from " ../generate.py --prefix=Ints -DELEM=[]int --package=tests --output=int_freepool.go ../randomized_freepool.go.tpl ". DO NOT EDIT.
// IntsFreePool is thread-safe pool that uses power-of-two loadbalancing across
// CPUs.

// This library requires the following two additional files per package. For
// now, create them manually.
//
// 1. A go file with the following contents
//
// package tests
//
// // This import is needed to use go:linkname.
// import _ "unsafe"
// // The following functions are defined in go runtime.  To use them, we need to
// // import "unsafe", and elsewhere in this package, import "C" to force compiler
// // to recognize the "go:linktime" directive. Some of the details are explained
// // in the below blog post.
// //
// // procPin() pins the caller to the current processor, and returns the processor
// // id in range [0,GOMAXPROCS). procUnpin() undos the effect of procPin().
// //
// // http://www.alangpierce.com/blog/2016/03/17/adventures-in-go-accessing-unexported-functions/
//
// //go:linkname runtime_procPin sync.runtime_procPin
// //go:nosplit
// func runtime_procPin() int
//
// //go:linkname runtime_procUnpin sync.runtime_procUnpin
// //go:nosplit
// func runtime_procUnpin()
//
// //go:linkname fastrand sync.fastrand
// func fastrand() uint32
//
// 2. An empty .s file.

package tests

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// IntsFreePool is a variation of sync.Pool, specialized for a concrete type.
//
// - Put() performs power-of-two loadbalancing, and Get() looks only at the
//   local queue.  This improves the performance of Get() on many-core machines,
//   at the cost of slightly more allocations.
//
// - It assumes that GOMAXPROCS is fixed at boot.
//
// - It never frees objects accumulated in the pool. We could add this feature
//   if needed.
type IntsFreePool struct {
	new          func() []int
	local        []IntspoolLocal
	maxLocalSize int64
}

const (
	IntsmaxPrivateElems = 4
	IntscacheLineSize   = 64
)

type IntspoolLocalInner struct {
	private     [IntsmaxPrivateElems][]int // Can be used only by the respective P.
	privateSize int

	shared     [][]int    // Can be used by any P.
	sharedSize int64      // ==len(shared), but can be accessed w/o holding mu.
	mu         sync.Mutex // Protects shared.
}

type IntspoolLocal struct {
	IntspoolLocalInner
	// Prevents false sharing.
	pad [IntscacheLineSize - unsafe.Sizeof(IntspoolLocalInner{})%IntscacheLineSize]byte
}

// NewIntsFreePool creates a new free object pool. new should create a new
// object. It is called when the pool is empty on Get(). maxSize bounds the
// approx max number of objects that can be stored in the pool. Beyond this
// limit, Put() call will drop the objects.
func NewIntsFreePool(new func() []int, maxSize int) *IntsFreePool {
	maxProcs := runtime.GOMAXPROCS(0)
	maxLocalSize := -1
	if maxSize > 0 {
		maxLocalSize = maxSize / maxProcs
		if maxLocalSize <= 0 {
			maxLocalSize = 1
		}
	}
	p := &IntsFreePool{
		new:          new,
		local:        make([]IntspoolLocal, maxProcs),
		maxLocalSize: int64(maxLocalSize),
	}
	return p
}

func (p *IntsFreePool) pin() *IntspoolLocal {
	pid := runtime_procPin()
	if int(pid) >= len(p.local) {
		panic(pid)
	}
	return &p.local[pid]
}

// Put adds an object to the freepool. The caller shall not touch the object
// after the call.
func (p *IntsFreePool) Put(x []int) {
	done := false
	l := p.pin()
	if l.privateSize < IntsmaxPrivateElems {
		l.private[l.privateSize] = x
		l.privateSize++
		done = true
	}
	runtime_procUnpin()
	if !done {
		// Pick another random queue, then add x to the shorter one.
		// This policy ("power of two") reduces load imbalance across
		// queues to log(log(#queues)) .
		//
		// https://www.eecs.harvard.edu/~michaelm/postscripts/mythesis.pdf
		l2 := &p.local[int(fastrand())%len(p.local)]
		lSize := atomic.LoadInt64(&l.sharedSize)
		l2Size := atomic.LoadInt64(&l2.sharedSize)
		if l2Size < lSize {
			l = l2
		}
		l.mu.Lock()
		if p.maxLocalSize >= 0 && l.sharedSize < p.maxLocalSize {
			l.shared = append(l.shared, x)
			atomic.StoreInt64(&l.sharedSize, l.sharedSize+1) // Release store.
		}
		l.mu.Unlock()
	}
}

// Get removes an object from the freepool. If pool is empty, it calls the
// callback passed to NewFreePool.
func (p *IntsFreePool) Get() []int {
	l := p.pin()
	var x []int
	done := false
	if l.privateSize > 0 {
		l.privateSize--
		x = l.private[l.privateSize]
		var empty []int
		l.private[l.privateSize] = empty
		done = true
	}
	runtime_procUnpin()
	if done {
		return x
	}
	l.mu.Lock()
	last := len(l.shared) - 1
	if last >= 0 {
		x = l.shared[last]
		l.shared = l.shared[:last]
		atomic.StoreInt64(&l.sharedSize, l.sharedSize-1)
		done = true
	}
	l.mu.Unlock()
	if !done {
		x = p.new()
	}
	return x
}

// ApproxLen returns an approximate length of the pool. For unittesting only.
//
// It returns an accurate value iff. no other thread is accessing the pool.
func (p *IntsFreePool) ApproxLen() int {
	n := 0
	for i := range p.local {
		n += p.local[i].privateSize
		n += int(p.local[i].sharedSize)
	}
	return n
}
