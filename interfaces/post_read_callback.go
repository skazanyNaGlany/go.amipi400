package interfaces

type PostReadCallback func(medium Medium, path string, buff []byte, ofst int64, fh uint64, n int, opTimeMs int64)
