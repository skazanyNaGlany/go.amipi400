package components

// from https://blog.zekro.de/hard-capping-buffer/

import (
	"bytes"
)

type CappedBuffer struct {
	*bytes.Buffer
	cap int
}

func New(buf []byte, cap int) *CappedBuffer {
	return &CappedBuffer{
		Buffer: bytes.NewBuffer(buf),
		cap:    cap,
	}
}

func (cb *CappedBuffer) Write(p []byte) (n int, err error) {
	if cb.cap > 0 && cb.Len()+len(p) > cb.cap {
		cb.Buffer.Reset()
	}

	return cb.Buffer.Write(p)
}

/*
func main() {
    buf := New(make([]byte, 0, 10), 10)

    fmt.Println(buf.Write([]byte{1, 2, 3, 4, 5}))
    fmt.Println(buf.Write([]byte{6, 7, 8, 9, 10, 11}))
}
*/
