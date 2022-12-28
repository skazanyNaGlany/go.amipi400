package medium

type FloppyMedium struct {
	MediumBase

	fullyCached       bool
	lastCachingTime   int64
	cachingNow        bool
	cachedAdfPathname string
	cachedAdfSha512   string
}

func (fm *FloppyMedium) SetCachedAdfPathname(cachedAdfPathname string) {
	fm.cachedAdfPathname = cachedAdfPathname
}

func (fm *FloppyMedium) GetCachedAdfPathname() string {
	return fm.cachedAdfPathname
}

func (fm *FloppyMedium) SetCachedAdfSha512(cachedAdfSha512 string) {
	fm.cachedAdfSha512 = cachedAdfSha512
}

func (fm *FloppyMedium) GetCachedAdfSha512() string {
	return fm.cachedAdfSha512
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

func (fm *FloppyMedium) Read(path string, buff []byte, ofst int64, fh uint64) (int, error) {
	return fm.driver.Read(fm, path, buff, ofst, fh)
}

func (mb *FloppyMedium) Write(path string, buff []byte, ofst int64, fh uint64) (int, error) {
	return mb.driver.Write(mb, path, buff, ofst, fh)
}
