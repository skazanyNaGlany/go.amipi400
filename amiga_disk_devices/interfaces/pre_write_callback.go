package interfaces

type PreWriteCallback func(medium Medium, path string, buff []byte, ofst int64, fh uint64)
