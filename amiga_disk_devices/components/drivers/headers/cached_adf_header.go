package headers

import (
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
)

type CachedADFHeader struct {
	Magic      [32]byte
	HeaderType [32]byte
	Sha512     [128]byte
	MTime      int64
}

func (cah *CachedADFHeader) Init() *CachedADFHeader {
	copy(cah.Magic[:], []byte(shared.CACHED_ADF_HEADER_MAGIC))
	copy(cah.HeaderType[:], []byte(shared.CACHED_ADF_HEADER_HEADER_TYPE))

	return cah
}

func (cah *CachedADFHeader) GetMagic() string {
	return strings.Trim(string(cah.Magic[:]), "\x00")
}

func (cah *CachedADFHeader) GetHeaderType() string {
	return strings.Trim(string(cah.HeaderType[:]), "\x00")
}

func (cah *CachedADFHeader) GetSha512() string {
	return strings.Trim(string(cah.Sha512[:]), "\x00")
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

func (cah *CachedADFHeader) IsValid() bool {
	magic := cah.GetMagic()
	headerType := cah.GetHeaderType()
	sha512 := cah.GetSha512()

	if magic != shared.CACHED_ADF_HEADER_MAGIC ||
		headerType != shared.CACHED_ADF_HEADER_HEADER_TYPE ||
		len(sha512) != shared.CACHED_ADF_HEADER_SHA512_LENGTH {
		return false
	}

	return utils.StringUtilsInstance.IsASCII(sha512)
}
