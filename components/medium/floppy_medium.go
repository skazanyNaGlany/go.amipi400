package medium

type FloppyMedium struct {
	MediumBase

	fullyCached     bool
	lastCachingTime int64
	cachingNow      bool
}

func (fm *FloppyMedium) IsFullyCached() bool {
	return fm.fullyCached
}

func (fm *FloppyMedium) SetFullyCached(fullyCached bool) {
	fm.fullyCached = fullyCached
}

func (fm *FloppyMedium) SetLastCachingTime(lastCachingTime int64) {
	fm.lastCachingTime = lastCachingTime
}

func (fm *FloppyMedium) GetLastCachingTime() int64 {
	return fm.lastCachingTime
}

func (fm *FloppyMedium) IsCachingNow() bool {
	return fm.cachingNow
}

func (fm *FloppyMedium) SetCachingNow(cachingNow bool) {
	fm.cachingNow = cachingNow
}

func (fm *FloppyMedium) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return fm.driver.Read(fm, path, buff, ofst, fh)
}

func (mb *FloppyMedium) Write(path string, buff []byte, ofst int64, fh uint64) int {
	return mb.driver.Write(mb, path, buff, ofst, fh)
}
