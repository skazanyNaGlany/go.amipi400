package consts

import "strings"

// global
const AMIPI400_UNIXNAME = "amipi400"
const AMIPI400_VERSION = "0.1"

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
const FLOPPY_MUTE_SOUND_NON_CACHED_READ = true
const FLOPPY_MUTE_SOUND_NON_CACHED_WRITE = true
const RUNNERS_VERBOSE_MODE = true
const RUNNERS_DEBUG_MODE = true
const DRIVERS_VERBOSE_MODE = true
const DRIVERS_DEBUG_MODE = true

var FORCE_INSERT_KEYS []string = []string{"L_SHIFT"}

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

var CACHED_ADF_HEADER_MAGIC = strings.ToUpper(AMIPI400_UNIXNAME + " v." + AMIPI400_VERSION)
