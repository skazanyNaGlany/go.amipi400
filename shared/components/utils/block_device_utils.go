package utils

import (
	"log"
	"strconv"
	"strings"

	"github.com/skazanyNaGlany/go.amipi400/shared"
)

type BlockDeviceUtils struct{}

var BlockDeviceUtilsInstance BlockDeviceUtils

func (bdu *BlockDeviceUtils) IsInternalMedium(name string) bool {
	return strings.HasPrefix(name, shared.SYSTEM_INTERNAL_SD_CARD_NAME)
}

func (bdu *BlockDeviceUtils) IsPoolMedium(name string) bool {
	return strings.HasPrefix(name, shared.POOL_DEVICE_NAME)
}

func (bdu *BlockDeviceUtils) PrintBlockDevice(
	name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) {
	log.Println("\tName:          " + name)
	log.Println("\tSize:          " + strconv.FormatUint(size, 10))
	log.Println("\tType:          " + _type)
	log.Println("\tMountpoint:    " + mountpoint)
	log.Println("\tLabel:         " + label)
	log.Println("\tPathname:      " + path)
	log.Println("\tFsType:        " + fsType)
	log.Println("\tPtType:        " + ptType)
	log.Println("\tRead-only:     " + strconv.FormatBool(readOnly))
}
