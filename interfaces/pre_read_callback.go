package interfaces

type PreReadCallback func(medium Medium, path string, buff []byte, ofst int64, fh uint64)
