package medium

import "os"

type FloppyMedium struct {
	MediumBase

	fullyCached          bool
	lastCachingTime      int64
	cachingNow           bool
	cachedAdfPathname    string
	cachedAdfSha512      string
	floppyUUID           string
	cachingDisabled      bool
	deviceDirectIOHandle *os.File
}

func (fm *FloppyMedium) GetDeviceDirectIOHandle() (*os.File, error) {
	return fm.deviceDirectIOHandle, nil
}

func (fm *FloppyMedium) SetDeviceDirectIOHandle(handle *os.File) {
	fm.deviceDirectIOHandle = handle
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

func (fm *FloppyMedium) SetFloppyUUID(uuid string) {
	fm.floppyUUID = uuid
}

func (fm *FloppyMedium) GetFloppyUUID() string {
	return fm.floppyUUID
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

func (fm *FloppyMedium) SetCachingDisabled(cachingDisabled bool) {
	fm.cachingDisabled = cachingDisabled
}

func (fm *FloppyMedium) IsCachingDisabled() bool {
	return fm.cachingDisabled
}

func (fm *FloppyMedium) Read(
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	return fm.driver.Read(fm, path, buff, ofst, fh)
}

func (mb *FloppyMedium) Write(
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	return mb.driver.Write(mb, path, buff, ofst, fh)
}

func (fm *FloppyMedium) Close() error {
	err := fm.driver.CloseMedium(fm)

	fm.CallClosedCallbacks(fm, err)

	return err
}
