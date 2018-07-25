// Code generated from " ../generate.py --prefix=byte -DMAXSIZE=128 -DELEM=[]byte --package=tests --output=freepool.go ../freepool.go.tpl ". DO NOT EDIT.
package tests

// A freepool for a single thread. The interface is the same as sync.Pool, but
// it avoids locks and interface conversion overhead.
//
// Example:
//  generate.py -package=foo -prefix=int -D[]byte=foo -D128=128
//
//
// Parameters:
//  []byte: the object to be kept in the freepool
//  128: the maxium number of objects to keep in the freepool

type bytePool struct {
	New func() []byte
	p   [][]byte
}

func (p *bytePool) Get() []byte {
	if len(p.p) == 0 {
		return p.New()
	}
	tmp := p.p[len(p.p)-1]
	p.p = p.p[:len(p.p)-1]
	return tmp
}

func (p *bytePool) Put(tmp []byte) {
	if len(p.p) < 128 {
		p.p = append(p.p, tmp)
	}
}
