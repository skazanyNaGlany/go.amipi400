package consts

import (
	"path/filepath"
	"strings"
)

// consts.go
const AMIPI400_UNIXNAME = "amipi400"
const AMIPI400_VERSION = "0.1"

// amipi400.go
const _AMIPI400_AMIBERRY_CONFIG_PATHNAME = "/boot/amipi400.uae.template"
const _AMIBERRY_EXE_PATHNAME = "../amiberry/amiberry"
const AMIBERRY_EMULATOR_TMP_INI_FILENAME = "amiberry.tmp.ini"

var SOFT_RESET_KEYS []string = []string{"L_CTRL", "L_ALT", "R_ALT"}

// amiga_disk_devices.go
const AMIGA_DISK_DEVICES_UNIXNAME = "amiga_disk_devices"
const AMIGA_DISK_DEVICES_VERSION = "0.1"
const SYSTEM_INTERNAL_SD_CARD_NAME = "mmcblk0"
const POOL_DEVICE_NAME = "loop"
const FILE_SYSTEM_MOUNT = "/tmp/amiga_disk_devices"
const CACHED_ADFS = "./cached_adfs"
const CACHED_ADFS_QUOTA = FLOPPY_ADF_SIZE * 1024 // 1024 adf files
const FLOPPY_READ_MUTE_SECS = 4
const FLOPPY_WRITE_MUTE_SECS = 4
const FLOPPY_WRITE_BLINK_POWER_SECS = 6
const HARD_DISK_READ_BLINK_POWER_SECS = 1
const FLOPPY_MUTE_SOUND_NON_CACHED_READ = true
const FLOPPY_MUTE_SOUND_NON_CACHED_WRITE = true
const RUNNERS_VERBOSE_MODE = true
const RUNNERS_DEBUG_MODE = true
const DRIVERS_VERBOSE_MODE = true
const DRIVERS_DEBUG_MODE = true

var FORCE_INSERT_KEYS []string = []string{"KEY_LEFTMETA", "L_SHIFT"}
var FORMAT_DEVICE_KEYS []string = []string{"KEY_LEFTMETA", "Del"}
var EMPTY_DEVICE_HEADER [2048]byte = [2048]byte{'D', 'O', 'S'}

// LEDControl
const LED0_BRIGHTNESS_PATHNAME = "/sys/class/leds/led0/brightness"

// CDMediumDriver
const CD_ISO_EXTENSION = "iso"
const CD_DEVICE_TYPE = "rom"
const CD_DEVICE_SECTOR_SIZE = 2048

// FloppyMediumDriver
const FLOPPY_DEVICE_SIZE = 1474560
const FLOPPY_ADF_SIZE = 901120
const FLOPPY_ADF_EXTENSION = "adf"
const FLOPPY_DEVICE_TYPE = "disk"
const FLOPPY_READ_AHEAD = 16
const FLOPPY_SECTOR_READ_TIME_MS = int64(100)
const FLOPPY_CACHE_DATA_BETWEEN_SECS = 3
const FLOPPY_DEVICE_SECTOR_SIZE = 512
const FLOPPY_DEVICE_LAST_SECTOR = 1474048

// HardDiskMediumDriver
const HD_DEVICE_MIN_SIZE = FLOPPY_DEVICE_SIZE + 1
const HD_HDF_EXTENSION = "hdf"
const HD_DEVICE_TYPE = "disk"
const HD_DEVICE_SECTOR_SIZE = 512

// MediumDriverBase
const DEFAULT_READ_AHEAD = 256

// AsyncFileOps
const ASYNC_FILE_OP_DIRECT_READ = "direct_read"
const ASYNC_FILE_OP_WRITE = "write"

// CachedADFHeader
const CACHED_ADF_HEADER_HEADER_TYPE = "CachedADFHeader"
const CACHED_ADF_HEADER_SHA512_LENGTH = 128

// amiga_disk_devices.go [2]
const HD_HDF_FULL_EXTENSION = "." + HD_HDF_EXTENSION

// CachedADFHeader [2]
var CACHED_ADF_HEADER_MAGIC = strings.ToUpper(AMIPI400_UNIXNAME + " v." + AMIPI400_VERSION)

// amipi400.go [2]
var AMIBERRY_EXE_PATHNAME, _ = filepath.Abs(_AMIBERRY_EXE_PATHNAME)
var AMIPI400_AMIBERRY_CONFIG_PATHNAME, _ = filepath.Abs(_AMIPI400_AMIBERRY_CONFIG_PATHNAME)
var AMIBERRY_EMULATOR_TMP_INI_PATHNAME = filepath.Join(
	filepath.Dir(AMIBERRY_EXE_PATHNAME),
	AMIBERRY_EMULATOR_TMP_INI_FILENAME)

// amipi400.go, EmulatorBase
const MAX_ADFS = 4
const FLOPPY_ADF_FULL_EXTENSION = "." + FLOPPY_ADF_EXTENSION

// AmiberryEmulator
const AMIBERRY_TEMPORARY_CONFIG_FILENAME = "amipi400.uae"
const OUTPUT_BUFFER_MAX_SIZE = 10485760
