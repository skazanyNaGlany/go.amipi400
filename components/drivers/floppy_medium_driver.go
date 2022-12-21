package drivers

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/skazanyNaGlany/go.amipi400/components"
	"github.com/skazanyNaGlany/go.amipi400/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/interfaces"
	"github.com/winfsp/cgofuse/fuse"
)

const floppyDeviceSize = 1474560
const floppyAdfSize = 901120
const floppyAdfExtension = "adf"
const floppyDeviceType = "disk"
const floppyReadAhead = 24
const floppySectorReadTimeMs = int64(100)
const floppyCacheDataBetweenSecs = 3
const floppyDeviceSectorSize = 512

var goUtils components.GoUtils

type FloppyMediumDriver struct {
	MediumDriverBase
}

func (fmd *FloppyMediumDriver) Probe(
	basePath, name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly bool) (interfaces.Medium, error) {
	// ignore medium which has MBR, or other known header
	// or known file-system or partition type, or just a label
	// detected by the system
	// Amiga ADF file is not known to the system
	// some games like Pinball Dreams Disc 2 has no valid DOS
	// header, but it is valid ADF file for the emulator
	// so we can use only these mediums which are unknown to the
	// system
	if fmd.isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
		return nil, nil
	}

	if size != floppyDeviceSize {
		return nil, nil
	}

	if _type != floppyDeviceType {
		return nil, nil
	}

	filename := strings.ReplaceAll(
		path,
		"/",
		"__")
	filename = filename + "." + floppyAdfExtension

	medium := medium.FloppyMedium{}

	medium.SetDriver(fmd)
	medium.SetDevicePathname(path)
	medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)
	medium.SetSize(floppyAdfSize)

	// in Linux all devices are readable by default
	medium.SetReadable(true)
	medium.SetWritable(!readOnly)

	now := time.Now().Unix()

	medium.SetCreateTime(now)
	medium.SetAccessTime(now)
	medium.SetModificationTime(now)

	return &medium, nil
}

func (mdb *FloppyMediumDriver) Read(_medium interfaces.Medium, path string, buff []byte, ofst int64, fh uint64) (n int) {
	mutex := _medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	_medium.SetAccessTime(
		time.Now().Unix())

	lenBuff := int64(len(buff))
	toReadSize := lenBuff
	fileSize := _medium.GetSize()

	if ofst+int64(toReadSize) > int64(fileSize) {
		toReadSize = fileSize - ofst
	}

	floppyMedium, castOk := _medium.(*medium.FloppyMedium)

	if !castOk {
		return -fuse.EIO
	}

	data, n_int64, err := mdb.read(floppyMedium, ofst, toReadSize)

	if err != nil {
		return -fuse.EIO
	}

	copy(buff, data)

	return int(n_int64)
}

func (mdb *FloppyMediumDriver) read(medium *medium.FloppyMedium, offset int64, size int64) ([]byte, int64, error) {
	// "rr" stand for "read_result"
	_floppyAdfSize := int64(floppyAdfSize)
	_floppySectorReadTimeMs := int64(floppySectorReadTimeMs)
	currentTime := time.Now().UnixMilli()

	if medium.GetLastCachingTime() == 0 {
		medium.SetLastCachingTime(currentTime)
	}

	rr_all_data, rr_total_read_time_ms, _, rr_err := mdb.partialRead(medium, offset, size, nil, nil)

	if rr_total_read_time_ms > floppySectorReadTimeMs {
		medium.SetFullyCached(false)
	}

	if medium.IsFullyCached() {
		// TODO spin
		return rr_all_data, int64(len(rr_all_data)), rr_err
	}

	if rr_total_read_time_ms < floppySectorReadTimeMs && !medium.IsFullyCached() &&
		currentTime-medium.GetLastCachingTime() >= floppyCacheDataBetweenSecs {

		medium.SetCachingNow(true)

		_, rr2_total_read_time_ms, _, _ := mdb.partialRead(
			medium,
			0,
			floppyDeviceSectorSize,
			&_floppyAdfSize,
			&_floppySectorReadTimeMs)

		medium.SetCachingNow(false)
		medium.SetLastCachingTime(currentTime)

		if rr2_total_read_time_ms < floppySectorReadTimeMs {
			medium.SetFullyCached(true)
		}
	}

	return rr_all_data, int64(len(rr_all_data)), rr_err
}

func (mdb *FloppyMediumDriver) partialRead(
	medium *medium.FloppyMedium,
	offset int64,
	size int64,
	max_read_size *int64,
	min_total_read_time_ms *int64) ([]byte, int64, int64, error) {
	all_data := make([]byte, 0)
	total_read_time_ms := int64(0)
	count_real_read_sectors := int64(0)
	to_read_size := size
	dynamic_offset := offset
	read_time_ms := int64(0)
	total_len_data := int64(0)

	handle, err := mdb.getMediumHandle(medium, floppyReadAhead)

	if err != nil {
		return nil, 0, 0, err
	}

	for {
		start_time := time.Now().UnixMilli()

		data, len_data, err := components.FileUtilsInstance.FileReadBytes(
			"",
			dynamic_offset,
			floppyDeviceSectorSize,
			0,
			0,
			handle)

		if err != nil {
			return data, 0, 0, err
		}

		dynamic_offset += int64(len_data)
		total_len_data += int64(len_data)

		read_time_ms = time.Now().UnixMilli() - start_time
		total_read_time_ms += read_time_ms

		if read_time_ms > floppySectorReadTimeMs {
			count_real_read_sectors += 1
		}

		all_data = append(all_data, data...)
		to_read_size -= int64(len_data)

		if len_data < floppyDeviceSectorSize {
			break
		}

		if max_read_size != nil {
			if total_len_data >= *max_read_size {
				break
			}
		}

		if to_read_size <= 0 {
			if min_total_read_time_ms != nil {
				if total_read_time_ms < *min_total_read_time_ms {
					continue
				}
			}

			break
		}
	}

	all_data = all_data[:size]

	return all_data, total_read_time_ms, count_real_read_sectors, nil
}
