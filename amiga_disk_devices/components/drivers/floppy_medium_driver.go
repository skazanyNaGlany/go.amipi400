package drivers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/ncw/directio"
	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/drivers/headers"
	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/components/medium"
	"github.com/skazanyNaGlany/go.amipi400/amiga_disk_devices/interfaces"
	"github.com/skazanyNaGlany/go.amipi400/shared"
	"github.com/skazanyNaGlany/go.amipi400/shared/components/utils"
	"github.com/winfsp/cgofuse/fuse"
)

type FloppyMediumDriver struct {
	MediumDriverBase

	cachedAdfsDirectory            string
	outsideAsyncFileWriterCallback interfaces.OutsideAsyncFileWriterCallback
	preCacheADFCallback            interfaces.PreCacheADFCallback
}

func (fmd *FloppyMediumDriver) Probe(
	basePath, name string,
	size uint64,
	_type, mountpoint, label, path, fsType, ptType string,
	readOnly, force, formatted bool) (interfaces.Medium, error) {
	// ignore medium which has MBR, or other known header
	// or known file-system or partition type, or just a label
	// detected by the system
	// Amiga ADF file is not known to the system
	// some games like Pinball Dreams Disc 2 has no valid DOS
	// header, but it is valid ADF file for the emulator
	// so we can use only these mediums which are unknown to the
	// system
	if !force {
		if fmd.isKnownMedium(name, mountpoint, label, path, fsType, ptType) {
			return nil, nil
		}
	}

	if size != shared.FLOPPY_DEVICE_SIZE {
		return nil, nil
	}

	if _type != shared.FLOPPY_DEVICE_TYPE {
		return nil, nil
	}

	_medium := medium.FloppyMedium{}

	filename := _medium.DevicePathnameToPublicFilename(path, shared.FLOPPY_ADF_EXTENSION)

	_medium.SetDriver(fmd)
	_medium.SetDevicePathname(path)
	_medium.SetPublicPathname(
		filepath.Join(basePath, filename),
	)
	_medium.SetSize(shared.FLOPPY_ADF_SIZE)

	// in Linux all devices are readable by default
	_medium.SetReadable(true)
	_medium.SetWritable(!readOnly)

	now := time.Now().Unix()

	_medium.SetCreateTime(now)
	_medium.SetAccessTime(now)
	_medium.SetModificationTime(now)

	// fail or not, we will re-cache the ADF again
	// if needed, sha512Id and UUID of the ADF should be set
	// in the FloppyMedium properly if it is known
	// (eg. ADF has its ID already, but cached ADF
	// does not exists)
	fmd.decodeCachedADFHeader(&_medium)

	if formatted || force {
		if _medium.GetCachedAdfPathname() != "" {
			// ADF is cached but the medium was
			// formatted, remove cached ADF file
			os.Remove(_medium.GetCachedAdfPathname())

			_medium.SetCachedAdfPathname("")

			if formatted {
				log.Println(
					"Removed cached ADF since the medium in",
					_medium.GetDevicePathname(),
					"was formatted",
				)
			}

			if force {
				log.Println(
					"Removed cached ADF since the medium in",
					_medium.GetDevicePathname(),
					"was forced to insert",
				)
			}
		}
	}

	return &_medium, nil
}

func (fmd *FloppyMediumDriver) floppyCacheAdf(_medium *medium.FloppyMedium) error {
	var sha512Id string
	var uuidStr string
	var err error
	var n int

	handle, err := fmd.OpenMediumHandle(_medium, shared.FLOPPY_READ_AHEAD)

	if err != nil {
		return err
	}

	data, len_data, err := utils.FileUtilsInstance.FileReadBytes(
		"",
		0,
		shared.FLOPPY_ADF_SIZE,
		0,
		0,
		handle)

	if err != nil {
		return err
	}

	if len_data < shared.FLOPPY_ADF_SIZE {
		return errors.New("cannot read medium data")
	}

	if len(data) < shared.FLOPPY_ADF_SIZE {
		return errors.New("cannot read medium data")
	}

	sha512Id = _medium.GetCachedAdfSha512()

	if sha512Id == "" {
		sha512Id = utils.CryptoUtilsInstance.BytesToSha512Hex(data)

		_medium.SetCachedAdfSha512(sha512Id)
	}

	uuidStr = _medium.GetFloppyUUID()

	if uuidStr == "" {
		// remove all - from the UUID
		uuidStr = strings.ReplaceAll(uuid.NewString(), "-", "")

		_medium.SetFloppyUUID(uuidStr)
	}

	cachedAdfPathname := path.Join(
		fmd.cachedAdfsDirectory,
		fmd.buildCachedAdfFilename(uuidStr, shared.FLOPPY_ADF_EXTENSION))

	if err = fmd.preCacheADFCallback(_medium, cachedAdfPathname); err != nil {
		return err
	}

	n, err = utils.FileUtilsInstance.FileWriteBytes(
		cachedAdfPathname,
		0,
		data,
		os.O_CREATE|os.O_WRONLY,
		0777,
		nil)

	if err != nil {
		return err
	}

	if n < len(data) {
		return errors.New("cannot create cached ADF file")
	}

	stat, _ := os.Stat(cachedAdfPathname)

	err = fmd.updateCachedADFHeader(
		_medium.GetDevicePathname(),
		sha512Id,
		uuidStr,
		stat.ModTime().Unix())

	if err != nil {
		return err
	}

	_medium.SetCachedAdfPathname(cachedAdfPathname)

	// close the handle, it will be opened
	// again for cached ADF
	if err = handle.Close(); err != nil {
		return err
	}

	_medium.SetHandle(nil)

	if fmd.verboseMode {
		log.Printf("ADF in medium %v have been cached\n", _medium.GetDevicePathname())
		log.Printf("\tCached ADF: %v\n", cachedAdfPathname)
		log.Printf("\tSHA512 ID:  %v\n", sha512Id)
		log.Printf("\tUUID:  %v\n", uuidStr)
	}

	return nil
}

func (fmd *FloppyMediumDriver) updateCachedADFHeader(
	pathname, sha512Id string,
	uuidStr string,
	mTime int64,
) error {
	header := new(headers.CachedADFHeader).Init()

	header.SetSha512(sha512Id)
	header.SetUUID(uuidStr)
	header.SetMTime(mTime)

	data, err := utils.GoUtilsInstance.StructToByteSlice(header)

	if err != nil {
		return err
	}

	fmd.outsideAsyncFileWriterCallback(
		pathname,
		shared.FLOPPY_DEVICE_LAST_SECTOR,
		data,
		os.O_SYNC|os.O_RDWR,
		0777,
		nil,
		true)

	return nil
}

func (fmd *FloppyMediumDriver) decodeCachedADFHeader(_medium *medium.FloppyMedium) error {
	header := headers.CachedADFHeader{}
	headerSize := unsafe.Sizeof(header)

	deviceRawHeader, n, err := utils.FileUtilsInstance.FileReadBytes(
		_medium.GetDevicePathname(),
		shared.FLOPPY_DEVICE_LAST_SECTOR,
		shared.FLOPPY_DEVICE_SECTOR_SIZE,
		os.O_RDONLY,
		0,
		nil)

	if err != nil {
		return err
	}

	if n < int(headerSize) {
		return fmt.Errorf("cannot read device's data, FileReadBytes returns %v", n)
	}

	if err = utils.GoUtilsInstance.ByteSliceToStruct(deviceRawHeader, &header); err != nil {
		return err
	}

	if !header.IsValid() {
		// CachedADFHeader is invalid or does not exists
		return nil
	}

	sha512Id := header.GetSha512()
	uuidStr := header.GetUUID()

	_medium.SetCachedAdfSha512(sha512Id)
	_medium.SetFloppyUUID(uuidStr)

	cachedAdfPathname := path.Join(
		fmd.cachedAdfsDirectory,
		fmd.buildCachedAdfFilename(uuidStr, shared.FLOPPY_ADF_EXTENSION))

	stat, err := os.Stat(cachedAdfPathname)

	if err != nil {
		return err
	}

	if stat.IsDir() {
		return errors.New("cached ADF file is a directory")
	}

	if stat.Size() < shared.FLOPPY_ADF_SIZE {
		return nil
	}

	if stat.ModTime().Unix() < header.GetMTime() {
		// cached ADF is older than physical medium
		// we will re-cache it later
		return nil
	}

	// it seems that ADF is properly cached
	_medium.SetCachedAdfPathname(cachedAdfPathname)

	if fmd.verboseMode {
		log.Printf("ADF in medium %v is cached\n", _medium.GetDevicePathname())
		log.Printf("\tCached ADF: %v\n", cachedAdfPathname)
		log.Printf("\tSHA512 ID:  %v\n", sha512Id)
		log.Printf("\tUUID:  %v\n", uuidStr)
	}

	return nil
}

func (fmd *FloppyMediumDriver) buildCachedAdfFilename(uuidStr, extension string) string {
	return uuidStr + "." + extension
}

func (mdb *FloppyMediumDriver) SetCachedAdfsDirectory(cachedAdfsDirectory string) {
	mdb.cachedAdfsDirectory = cachedAdfsDirectory
}

func (fmd *FloppyMediumDriver) OpenMediumHandle(
	_medium interfaces.Medium,
	readAhead ...int,
) (*os.File, error) {
	floppyMedium, castOk := _medium.(*medium.FloppyMedium)

	if !castOk {
		return nil, errors.New("cannot cast Medium to FloppyMedium")
	}

	handle, err := floppyMedium.GetHandle()

	if err != nil {
		return nil, err
	}

	if handle != nil {
		// already open
		return handle, nil
	}

	isReadable := floppyMedium.IsReadable()
	isWritable := floppyMedium.IsWritable()

	flag := os.O_SYNC

	if isReadable && isWritable {
		flag |= os.O_RDWR
	} else {
		flag |= os.O_RDONLY
	}

	cachedAdfPathname := floppyMedium.GetCachedAdfPathname()
	devicePathname := floppyMedium.GetDevicePathname()

	if cachedAdfPathname == "" && devicePathname != "" {
		// not cached yet, just open device handle
		return fmd.openDeviceHandle(floppyMedium, flag, readAhead...)
	} else if cachedAdfPathname != "" && devicePathname != "" {
		// cached, open two handles, one for device, second for cached ADF
		// set 0 bytes read-a-head for device handle to speed-up FileReadBytesDirect
		// (to speed-up moving the motor)
		return fmd.openDeviceAndADFHandles(floppyMedium, flag, readAhead...)
	}

	return nil, nil
}

func (fmd *FloppyMediumDriver) openDeviceHandle(
	floppyMedium *medium.FloppyMedium,
	flag int,
	readAhead ...int,
) (*os.File, error) {
	devicePathname := floppyMedium.GetDevicePathname()

	deviceHandle, err := os.OpenFile(
		devicePathname,
		flag,
		0,
	)

	if err != nil {
		return nil, err
	}

	_readAhead := shared.DEFAULT_READ_AHEAD

	if len(readAhead) == 1 {
		_readAhead = readAhead[0]
	}

	if err = utils.UnixUtilsInstance.SetDeviceReadAHead(deviceHandle, _readAhead); err != nil {
		deviceHandle.Close()

		return nil, err
	}

	floppyMedium.SetHandle(deviceHandle)

	return deviceHandle, nil

}

func (fmd *FloppyMediumDriver) openDeviceAndADFHandles(
	floppyMedium *medium.FloppyMedium,
	flag int,
	readAhead ...int,
) (*os.File, error) {
	cachedAdfPathname := floppyMedium.GetCachedAdfPathname()
	devicePathname := floppyMedium.GetDevicePathname()

	// cached ADF handle, will be used as main handle for all read/write operations
	cachedAdfHandle, err := os.OpenFile(
		cachedAdfPathname,
		flag,
		0,
	)

	if err != nil {
		return nil, err
	}

	// device deviceHandle, will be used to move the motor
	// to move real floppy drive motor we need a special direct-io-handle
	// instead of a regular one
	deviceHandle, err := directio.OpenFile(devicePathname, flag, 0)

	if err != nil {
		cachedAdfHandle.Close()

		return nil, err
	}

	if err = utils.UnixUtilsInstance.SetDeviceReadAHead(deviceHandle, 0); err != nil {
		cachedAdfHandle.Close()
		deviceHandle.Close()

		return nil, err
	}

	floppyMedium.SetHandle(cachedAdfHandle)
	floppyMedium.SetDeviceDirectIOHandle(deviceHandle)

	return cachedAdfHandle, nil
}

func (fmd *FloppyMediumDriver) CloseMedium(_medium interfaces.Medium) error {
	floppyMedium, castOk := _medium.(*medium.FloppyMedium)

	if !castOk {
		return errors.New("cannot cast Medium to FloppyMedium")
	}

	handle, err0 := floppyMedium.GetHandle()
	directIOhandle, err1 := floppyMedium.GetDeviceDirectIOHandle()

	if handle == nil {
		// handle not open yet, or already closed
		return nil
	}

	floppyMedium.SetHandle(nil)
	floppyMedium.SetDeviceDirectIOHandle(nil)

	if err0 != nil {
		return err0
	}

	if err1 != nil {
		return err1
	}

	err0 = handle.Close()
	err1 = directIOhandle.Close()

	if err0 != nil {
		return err0
	}

	if err1 != nil {
		return err1
	}

	return nil
}

func (fmd *FloppyMediumDriver) SetOutsideAsyncFileWriterCallback(
	callback interfaces.OutsideAsyncFileWriterCallback,
) {
	fmd.outsideAsyncFileWriterCallback = callback
}

func (fmd *FloppyMediumDriver) SetPreCacheADFCallback(
	callback interfaces.PreCacheADFCallback,
) {
	fmd.preCacheADFCallback = callback
}

// Read
func (fmd *FloppyMediumDriver) Read(
	_medium interfaces.Medium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	mutex := _medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	floppyMedium, castOk := _medium.(*medium.FloppyMedium)

	if !castOk {
		return 0, errors.New("cannot cast Medium to FloppyMedium")
	}

	floppyMedium.SetAccessTime(
		time.Now().Unix())

	lenBuff := int64(len(buff))
	toReadSize := lenBuff
	fileSize := floppyMedium.GetSize()

	if ofst+int64(toReadSize) > int64(fileSize) {
		toReadSize = fileSize - ofst
	}

	if floppyMedium.GetCachedAdfPathname() == "" {
		return fmd.realRead(floppyMedium, path, buff, ofst, toReadSize, fh)
	}

	return fmd.cachedRead(floppyMedium, path, buff, ofst, toReadSize, fh)
}

func (fmd *FloppyMediumDriver) realRead(
	floppyMedium *medium.FloppyMedium,
	path string,
	buff []byte,
	ofst, toReadSize int64,
	fh uint64,
) (int, error) {
	data, n_int64, err := fmd.realRead2(floppyMedium, path, ofst, toReadSize, fh)

	if err != nil {
		return 0, err
	}

	copy(buff, data)

	return int(n_int64), nil
}

func (mdb *FloppyMediumDriver) realRead2(
	floppyMedium *medium.FloppyMedium,
	path string,
	offset int64,
	size int64,
	fh uint64) ([]byte, int64, error) {
	// "rr" stand for "read_result"
	_floppyAdfSize := int64(shared.FLOPPY_ADF_SIZE)
	_floppySectorReadTimeMs := int64(shared.FLOPPY_SECTOR_READ_TIME_MS)
	currentTime := time.Now().UnixMilli()

	if floppyMedium.GetLastCachingTime() == 0 {
		floppyMedium.SetLastCachingTime(currentTime)
	}

	rr_all_data, rr_total_read_time_ms, _, rr_err := mdb.partialRead(
		floppyMedium,
		path,
		offset,
		size,
		nil,
		nil,
		fh,
	)

	if rr_total_read_time_ms > shared.FLOPPY_SECTOR_READ_TIME_MS {
		floppyMedium.SetFullyCached(false)
	}

	if floppyMedium.IsFullyCached() {
		return rr_all_data, int64(len(rr_all_data)), rr_err
	}

	if rr_total_read_time_ms < shared.FLOPPY_SECTOR_READ_TIME_MS &&
		!floppyMedium.IsFullyCached() &&
		currentTime-floppyMedium.GetLastCachingTime() >= shared.FLOPPY_CACHE_DATA_BETWEEN_SECS {

		floppyMedium.SetCachingNow(true)

		_, rr2_total_read_time_ms, _, _ := mdb.partialRead(
			floppyMedium,
			path,
			0,
			shared.FLOPPY_DEVICE_SECTOR_SIZE,
			&_floppyAdfSize,
			&_floppySectorReadTimeMs,
			fh)

		floppyMedium.SetCachingNow(false)
		floppyMedium.SetLastCachingTime(currentTime)

		if rr2_total_read_time_ms < shared.FLOPPY_SECTOR_READ_TIME_MS {
			floppyMedium.SetFullyCached(true)

			if !floppyMedium.IsCachingDisabled() && floppyMedium.IsWritable() {
				err := mdb.floppyCacheAdf(floppyMedium)

				if err != nil {
					// cannot cache the ADF, disable
					// caching for that one
					floppyMedium.SetCachingDisabled(true)

					if mdb.debugMode {
						log.Println(err)
					}
				}
			} else {
				if mdb.verboseMode {
					log.Printf("Caching is disabled for medium in %v\n", floppyMedium.GetDevicePathname())
				}
			}
		}
	}

	return rr_all_data, int64(len(rr_all_data)), rr_err
}

func (mdb *FloppyMediumDriver) partialRead(
	medium *medium.FloppyMedium,
	path string,
	offset int64,
	size int64,
	max_read_size *int64,
	min_total_read_time_ms *int64,
	fh uint64) ([]byte, int64, int64, error) {
	all_data := make([]byte, 0)
	total_read_time_ms := int64(0)
	count_real_read_sectors := int64(0)
	to_read_size := size
	dynamic_offset := offset
	read_time_ms := int64(0)
	total_len_data := int64(0)

	handle, err := mdb.OpenMediumHandle(medium, shared.FLOPPY_READ_AHEAD)

	if err != nil {
		return nil, 0, 0, err
	}

	for {
		start_time := time.Now().UnixMilli()

		medium.CallPreReadCallbacks(
			medium,
			path,
			all_data,
			dynamic_offset,
			fh)

		data, len_data, err := utils.FileUtilsInstance.FileReadBytes(
			"",
			dynamic_offset,
			shared.FLOPPY_DEVICE_SECTOR_SIZE,
			0,
			0,
			handle)

		read_time_ms = time.Now().UnixMilli() - start_time
		total_read_time_ms += read_time_ms

		if err != nil {
			medium.CallPostReadCallbacks(
				medium,
				path,
				data,
				dynamic_offset,
				fh,
				-fuse.EIO,
				read_time_ms,
			)

			return data, 0, 0, err
		}

		dynamic_offset += int64(len_data)
		total_len_data += int64(len_data)

		medium.CallPostReadCallbacks(
			medium,
			path,
			data,
			dynamic_offset,
			fh,
			len_data,
			read_time_ms,
		)

		if read_time_ms > shared.FLOPPY_SECTOR_READ_TIME_MS {
			count_real_read_sectors += 1
		}

		all_data = append(all_data, data...)
		to_read_size -= int64(len_data)

		if len_data < shared.FLOPPY_DEVICE_SECTOR_SIZE {
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

func (fmd *FloppyMediumDriver) cachedRead(
	floppyMedium *medium.FloppyMedium,
	path string,
	buff []byte,
	ofst, toReadSize int64,
	fh uint64,
) (int, error) {
	handle, err := fmd.OpenMediumHandle(floppyMedium)

	if err != nil {
		return 0, err
	}

	floppyMedium.SetFullyCached(true)

	all_data := make([]byte, 0)

	floppyMedium.CallPreReadCallbacks(
		floppyMedium,
		path,
		all_data,
		ofst,
		fh)

	data, n, err := utils.FileUtilsInstance.FileReadBytes(
		"",
		ofst,
		toReadSize,
		0,
		0,
		handle)

	if err != nil {
		floppyMedium.CallPostReadCallbacks(
			floppyMedium,
			path,
			data,
			ofst,
			fh,
			-fuse.EIO,
			0,
		)

		return 0, err
	}

	floppyMedium.CallPostReadCallbacks(floppyMedium, path, data, ofst, fh, n, 0)

	copy(buff, data)

	return n, nil
}

// Write
func (fmd *FloppyMediumDriver) Write(
	_medium interfaces.Medium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	// Almost the same as MediumDriverBase.Write, but calling SetFullyCached also
	mutex := _medium.GetMutex()

	mutex.Lock()
	defer mutex.Unlock()

	if !_medium.IsWritable() {
		return -fuse.EPERM, errors.New("device is not writable")
	}

	floppyMedium, castOk := _medium.(*medium.FloppyMedium)

	if !castOk {
		return 0, errors.New("cannot cast Medium to FloppyMedium")
	}

	floppyMedium.SetModificationTime(
		time.Now().Unix())

	fileSize := floppyMedium.GetSize()
	lenBuff := len(buff)

	if ofst+int64(lenBuff) > fileSize || ofst >= fileSize {
		return 0, errors.New("Write outside the medium")
	}

	if floppyMedium.GetCachedAdfPathname() == "" {
		return fmd.realWrite(floppyMedium, path, buff, ofst, fh)
	}

	return fmd.cachedWrite(floppyMedium, path, buff, ofst, fh)
}

func (fmd *FloppyMediumDriver) realWrite(
	floppyMedium *medium.FloppyMedium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	handle, err := fmd.OpenMediumHandle(floppyMedium, shared.FLOPPY_READ_AHEAD)

	if err != nil {
		return 0, err
	}

	floppyMedium.SetFullyCached(false)
	floppyMedium.CallPreWriteCallbacks(floppyMedium, path, buff, ofst, fh)

	startTime := time.Now().UnixMilli()
	n, err := utils.FileUtilsInstance.FileWriteBytes("", ofst, buff, 0, 0, handle)
	totalTime := time.Now().UnixMilli() - startTime

	if err != nil {
		floppyMedium.CallPostWriteCallbacks(
			floppyMedium,
			path,
			buff,
			ofst,
			fh,
			-fuse.EIO,
			totalTime,
		)

		return 0, err
	}

	floppyMedium.CallPostWriteCallbacks(floppyMedium, path, buff, ofst, fh, n, totalTime)

	return n, nil
}

func (fmd *FloppyMediumDriver) cachedWrite(
	floppyMedium *medium.FloppyMedium,
	path string,
	buff []byte,
	ofst int64,
	fh uint64,
) (int, error) {
	handle, err := fmd.OpenMediumHandle(floppyMedium)

	if err != nil {
		return 0, err
	}

	floppyMedium.SetFullyCached(true)
	floppyMedium.CallPreWriteCallbacks(floppyMedium, path, buff, ofst, fh)

	n, err := utils.FileUtilsInstance.FileWriteBytes("", ofst, buff, 0, 0, handle)

	fmd.outsideAsyncFileWriterCallback(
		floppyMedium.GetDevicePathname(),
		ofst,
		buff,
		os.O_SYNC|os.O_RDWR,
		0777,
		nil,
		false)

	stat, _ := os.Stat(
		floppyMedium.GetCachedAdfPathname())

	fmd.updateCachedADFHeader(
		floppyMedium.GetDevicePathname(),
		floppyMedium.GetCachedAdfSha512(),
		floppyMedium.GetFloppyUUID(),
		stat.ModTime().Unix())

	if err != nil {
		floppyMedium.CallPostWriteCallbacks(
			floppyMedium,
			path,
			buff,
			ofst,
			fh,
			-fuse.EIO,
			0,
		)

		return 0, err
	}

	floppyMedium.CallPostWriteCallbacks(floppyMedium, path, buff, ofst, fh, n, 0)

	return n, nil
}
