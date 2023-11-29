package shared

import (
	"math"
	"path/filepath"
	"regexp"
	"strings"
)

// consts.go
const AMIPI400_UNIXNAME = "amipi400"
const AMIPI400_VERSION = "0.1"

// global
const HDD_SECTOR_SIZE = 512
const CMD_PENDING_BLINK_POWER_SECS = math.MaxInt
const CMD_FAILURE_BLINK_NUM_LOCK_SECS = 4

// amipi400.go
const _AMIPI400_AMIBERRY_CONFIG_PATHNAME = "/boot/amipi400.uae.template"
const _AMIBERRY_EXE_PATHNAME = "../../amiberry/amiberry"
const AMIBERRY_EMULATOR_TMP_INI_FILENAME = "amiberry.tmp.ini"
const AP4_ROOT_MOUNTPOINT = "/media/"
const FLOPPY_DISK_IN_DRIVE_SOUND_VOLUME = 25
const HARD_RESET_KEYS_MIN_MS = 4000 // min 4 seconds
const MEDIUM_CONFIG_INI_NAME = "amipi400.ini"
const MEDIUM_CONFIG_DEFAULT_SECTION = "amipi400"
const MEDIUM_CONFIG_DEFAULT_FILE = "default_file"
const MEDIUM_CONFIG_DEFAULT_FILE_NONE = "none"
const ADF_DISK_NO_OF_MAX = "(Disk %d of %d)"
const AUTORUN_EMULATOR = true

// TODO replace [0-9] with \d
var SOFT_RESET_KEYS []string = []string{"L_CTRL", "L_ALT", "R_ALT"}
var HARD_RESET_KEYS []string = []string{"L_CTRL", "L_ALT", "R_ALT"}
var AP4_MEDIUM_DF_RE = regexp.MustCompile(`^AP4_DF(?P<index>[0-9]?|X)$`)
var AP4_MEDIUM_DH_RE = regexp.MustCompile(`^AP4_DH(?P<index>[0-9]?|X)(_(?P<boot_priority>[0-9]))?$`)
var AP4_MEDIUM_HF_RE = regexp.MustCompile(`^AP4_HF(?P<index>[0-9]?|X)(_(?P<boot_priority>[0-9]))?$`)
var AP4_MEDIUM_CD_RE = regexp.MustCompile(`^AP4_CD(?P<index>[0-9]?|X)$`)
var DF_INSERT_FROM_SOURCE_TO_TARGET_INDEX_RE = regexp.MustCompile(`^DF(?P<source_index>[0-9])(?P<filename_part>.*)DF(?P<target_index>[0-9]|N)$`)
var DF_INSERT_FROM_SOURCE_TO_TARGET_INDEX_BY_DISK_NO_RE = regexp.MustCompile(`^DF(?P<source_index>[0-9])(?P<disk_no>\d\d?)DF(?P<target_index>[0-9]|N)$`)
var DF_INSERT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^DF(?P<source_index>[0-9])(?P<filename_part>.*)$`)
var DF_INSERT_FROM_SOURCE_INDEX_BY_DISK_NO_RE = regexp.MustCompile(`^DF(?P<source_index>[0-9])(?P<disk_no>\d\d?)$`)
var DF_EJECT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^DF(?P<source_index>[0-9]|N)$`)
var CD_INSERT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^CD(?P<source_index>[0-9])(?P<filename_part>.*)$`)
var CD_EJECT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^CD(?P<source_index>[0-9])$`)
var ADF_DISK_NO_OF_MAX_RE = regexp.MustCompile(`(?P<disk_no_of_max>\((Disk\ \d)\ (of\ \d)\))`)
var HF_INSERT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^HF(?P<source_index>[0-9])(?P<filename_part>.*)$`)
var HF_EJECT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^HF(?P<source_index>[0-9])$`)
var DF_UNMOUNT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^UDF(?P<source_index>\d|N)$`)
var CD_UNMOUNT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^UCD(?P<source_index>\d|N)$`)
var HF_UNMOUNT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^UHF(?P<source_index>\d|N)$`)
var DH_UNMOUNT_FROM_SOURCE_INDEX_RE = regexp.MustCompile(`^UDH(?P<source_index>\d|N)$`)
var UNMOUNT_ALL_RE = regexp.MustCompile(`^U$`)

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
const FLOPPY_WRITE_BLINK_POWER_SECS = 8
const HARD_DISK_READ_BLINK_POWER_SECS = 4
const FLOPPY_MUTE_SOUND_NON_CACHED_READ = true
const FLOPPY_MUTE_SOUND_NON_CACHED_WRITE = true
const RUNNERS_VERBOSE_MODE = true
const RUNNERS_DEBUG_MODE = true
const DRIVERS_VERBOSE_MODE = true
const DRIVERS_DEBUG_MODE = true

var FORCE_INSERT_KEYS []string = []string{"KEY_LEFTMETA", "L_SHIFT"}
var FORMAT_DEVICE_KEYS []string = []string{"KEY_LEFTMETA", "Del"}
var EMPTY_DEVICE_HEADER [2048]byte = [2048]byte{'D', 'O', 'S'}
var TOGGLE_AUTO_HEIGHT_KEYS []string = []string{"KEY_LEFTMETA", "H"}
var AMIGA_DISK_DEVICES_NEEDED_EXECUTABLES = []string{"sync", "fsck", "ufiformat", "hwinfo", "lsblk"}

// PowerLEDControl
const POWER_LED0_BRIGHTNESS_PATHNAME = "/sys/class/leds/led0/brightness"

// NumLockLEDControl
const NUM_LOCK_LED0_BRIGHTNESS_PATHNAME = "/sys/class/leds/input0::numlock/brightness"

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
var AMIPI400_NEEDED_EXECUTABLES = []string{"sync", "fsck", "ufiformat"}

const DRIVE_INDEX_UNSPECIFIED = -1
const DRIVE_INDEX_UNSPECIFIED_STR = "-1"
const DH_BOOT_PRIORITY_DEFAULT = 0
const DH_BOOT_PRIORITY_UNSPECIFIED = -1

// amipi400.go, AmiberryEmulator
const MAX_ADFS = 4
const FLOPPY_ADF_FULL_EXTENSION = "." + FLOPPY_ADF_EXTENSION
const MAX_HDFS = 7
const MAX_CDS = 1
const CD_ISO_FULL_EXTENSION = "." + CD_ISO_EXTENSION

// AmiberryEmulator
const AMIBERRY_TEMPORARY_CONFIG_FILENAME = "amipi400.uae"
const OUTPUT_BUFFER_MAX_SIZE = 10485760
const HDF_TYPE_HDFRDB = 8
const HDF_TYPE_DISKIMAGE = 2
const HDF_TYPE_HDF = 5
const AMIBERRY_DEFAULT_WINDOW_HEIGHT = 568
const AMIBERRY_ZOOM_WINDOW_HEIGHT = 512

// AllKeyboardsControl / KeyboardControl
const MAX_KEYS_SEQUENCE = 128
const KEY_ESC = "ESC"
const KEY_TAB = "TAB"
const KEY_SPACE = "SPACE"
