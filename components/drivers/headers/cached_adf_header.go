package headers

import (
	"strings"
)

var cachedADFHeaderHeaderType = "CachedADFHeader"
var cachedADFHeaderSHA512Length = 128

type CachedADFHeader struct {
	Magic      [32]byte
	HeaderType [32]byte
	Sha512     [128]byte
	MTime      int64
}

func (cah *CachedADFHeader) GetMagic() string {
	return strings.TrimRight(string(cah.Magic[:]), "\x00")
}

func (cah *CachedADFHeader) SetMagic(magic string) {
	copy(cah.Magic[:], []byte(magic))
}

func (cah *CachedADFHeader) GetHeaderType() string {
	return strings.TrimRight(string(cah.HeaderType[:]), "\x00")
}

func (cah *CachedADFHeader) SetHeaderType(headerType string) {
	copy(cah.HeaderType[:], []byte(headerType))
}

func (cah *CachedADFHeader) GetSha512() string {
	return strings.TrimRight(string(cah.Sha512[:]), "\x00")
}

func (cah *CachedADFHeader) SetSha512(sha512 string) {
	copy(cah.Sha512[:], []byte(sha512))
}

func (cah *CachedADFHeader) SetMTime(mTime int64) {
	cah.MTime = mTime
}

func (cah *CachedADFHeader) GetMTime() int64 {
	return cah.MTime
}

func (cah *CachedADFHeader) IsValid(magic string) bool {
	return cah.GetMagic() == magic &&
		cah.GetHeaderType() == cachedADFHeaderHeaderType &&
		len(cah.GetSha512()) == cachedADFHeaderSHA512Length
}
