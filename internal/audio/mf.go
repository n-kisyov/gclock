package audio

import (
	"syscall"
	"unsafe"
	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

var (
	modmfplat      = windows.NewLazySystemDLL("mfplat.dll")
	modmfreadwrite = windows.NewLazySystemDLL("mfreadwrite.dll")

	procMFStartup                   = modmfplat.NewProc("MFStartup")
	procMFShutdown                  = modmfplat.NewProc("MFShutdown")
	procMFCreateMediaType           = modmfplat.NewProc("MFCreateMediaType")
	procMFCreateSourceReaderFromURL = modmfreadwrite.NewProc("MFCreateSourceReaderFromURL")
)

const (
	MF_VERSION      = 0x00020070 // typically 0x00020070 for Windows 7+
	MFSTARTUP_LITE  = 1
)

var (
	MF_MT_MAJOR_TYPE                 = windows.GUID{Data1: 0x48EBA18E, Data2: 0xF8C9, Data3: 0x4687, Data4: [8]byte{0xBF, 0x11, 0x0A, 0x74, 0xC9, 0xF9, 0x6A, 0x8F}}
	MFMediaType_Audio                = windows.GUID{Data1: 0x73647561, Data2: 0x0000, Data3: 0x0010, Data4: [8]byte{0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71}}
	MF_MT_SUBTYPE                    = windows.GUID{Data1: 0xF7E34C9A, Data2: 0x42E8, Data3: 0x4714, Data4: [8]byte{0xB7, 0x4B, 0xCB, 0x29, 0xD7, 0x2C, 0x35, 0xE5}}
	MFAudioFormat_Float              = windows.GUID{Data1: 0x00000003, Data2: 0x0000, Data3: 0x0010, Data4: [8]byte{0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71}}
	MF_MT_AUDIO_SAMPLES_PER_SECOND   = windows.GUID{Data1: 0x5FAEEAE7, Data2: 0x0290, Data3: 0x4C31, Data4: [8]byte{0x9E, 0x8A, 0xC5, 0x34, 0xF6, 0x8D, 0x9D, 0xBA}}
	MF_MT_AUDIO_NUM_CHANNELS         = windows.GUID{Data1: 0x37E48BF5, Data2: 0x645E, Data3: 0x4C5B, Data4: [8]byte{0x89, 0xDE, 0xAD, 0xA9, 0xE2, 0x9B, 0x69, 0x6A}}
	MF_MT_AUDIO_BITS_PER_SAMPLE      = windows.GUID{Data1: 0xF2DEB57F, Data2: 0x40FA, Data3: 0x4764, Data4: [8]byte{0xAA, 0x33, 0xED, 0x4F, 0x2D, 0x1F, 0xF6, 0x69}}
	MF_MT_AUDIO_BLOCK_ALIGNMENT      = windows.GUID{Data1: 0x322DE230, Data2: 0x9EEB, Data3: 0x43BD, Data4: [8]byte{0xAB, 0x7A, 0xFF, 0x41, 0x22, 0x51, 0x54, 0x1D}}
	MF_MT_AUDIO_AVG_BYTES_PER_SECOND = windows.GUID{Data1: 0x1AAB75C8, Data2: 0xCFEF, Data3: 0x451C, Data4: [8]byte{0xAB, 0x95, 0xAC, 0x03, 0x4B, 0x8E, 0x17, 0x31}}
	MF_MT_ALL_SAMPLES_INDEPENDENT    = windows.GUID{Data1: 0xC9173739, Data2: 0x5E56, Data3: 0x461C, Data4: [8]byte{0xB7, 0x13, 0x46, 0xFB, 0x99, 0x5C, 0xB9, 0x5F}}
)

const (
	MF_SOURCE_READER_ALL_STREAMS        = 0xFFFFFFFE
	MF_SOURCE_READER_FIRST_AUDIO_STREAM = 0xFFFFFFFD
	MF_SOURCE_READERF_ENDOFSTREAM       = 0x00000002
)

// -- IMFSourceReader --
type IMFSourceReader struct { ole.IUnknown }
type IMFSourceReaderVtbl struct {
	ole.IUnknownVtbl
	GetStreamSelection   uintptr
	SetStreamSelection   uintptr
	GetNativeMediaType   uintptr
	GetCurrentMediaType  uintptr
	SetCurrentMediaType  uintptr
	SetCurrentPosition   uintptr
	ReadSample           uintptr
	Flush                uintptr
	GetServiceForStream  uintptr
	GetPresentationAttribute uintptr
}
func (v *IMFSourceReader) VTable() *IMFSourceReaderVtbl { return (*IMFSourceReaderVtbl)(unsafe.Pointer(v.RawVTable)) }
func (v *IMFSourceReader) SetStreamSelection(index uint32, selected bool) error {
	sel := uintptr(0)
	if selected { sel = 1 }
	ret, _, _ := syscall.Syscall(v.VTable().SetStreamSelection, 3, uintptr(unsafe.Pointer(v)), uintptr(index), sel)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func (v *IMFSourceReader) SetCurrentMediaType(index uint32, reserved uintptr, mt *IMFMediaType) error {
	ret, _, _ := syscall.Syscall6(v.VTable().SetCurrentMediaType, 4, uintptr(unsafe.Pointer(v)), uintptr(index), reserved, uintptr(unsafe.Pointer(mt)), 0, 0)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func (v *IMFSourceReader) ReadSample(index uint32, flags uint32, actualIndex *uint32, sampleFlags *uint32, timestamp *uint64, sample **IMFSample) error {
	ret, _, _ := syscall.Syscall9(v.VTable().ReadSample, 7, uintptr(unsafe.Pointer(v)), uintptr(index), uintptr(flags),
		uintptr(unsafe.Pointer(actualIndex)), uintptr(unsafe.Pointer(sampleFlags)), uintptr(unsafe.Pointer(timestamp)), uintptr(unsafe.Pointer(sample)), 0, 0)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}

// -- IMFMediaType (inherits IMFAttributes) --
type IMFMediaType struct { ole.IUnknown }
type IMFMediaTypeVtbl struct {
	ole.IUnknownVtbl
	GetItem             uintptr
	GetItemType         uintptr
	CompareItem         uintptr
	Compare             uintptr
	GetUINT32           uintptr
	GetUINT64           uintptr
	GetDouble           uintptr
	GetGUID             uintptr
	GetStringLength     uintptr
	GetString           uintptr
	GetAllocatedString  uintptr
	GetBlobSize         uintptr
	GetBlob             uintptr
	GetAllocatedBlob    uintptr
	GetUnknown          uintptr
	SetItem             uintptr
	DeleteItem          uintptr
	DeleteAllItems      uintptr
	SetUINT32           uintptr
	SetUINT64           uintptr
	SetDouble           uintptr
	SetGUID             uintptr
	SetString           uintptr
	SetBlob             uintptr
	SetUnknown          uintptr
	LockStore           uintptr
	UnlockStore         uintptr
	GetCount            uintptr
	GetItemByIndex      uintptr
	CopyAllItems        uintptr
	// IMFMediaType methods follow
	GetMajorType        uintptr
	IsCompressedFormat  uintptr
	IsEqual             uintptr
	GetRepresentation   uintptr
	FreeRepresentation  uintptr
}
func (v *IMFMediaType) VTable() *IMFMediaTypeVtbl { return (*IMFMediaTypeVtbl)(unsafe.Pointer(v.RawVTable)) }
func (v *IMFMediaType) SetGUID(guidKey *windows.GUID, guidValue *windows.GUID) error {
	ret, _, _ := syscall.Syscall(v.VTable().SetGUID, 3, uintptr(unsafe.Pointer(v)), uintptr(unsafe.Pointer(guidKey)), uintptr(unsafe.Pointer(guidValue)))
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func (v *IMFMediaType) SetUINT32(guidKey *windows.GUID, value uint32) error {
	ret, _, _ := syscall.Syscall(v.VTable().SetUINT32, 3, uintptr(unsafe.Pointer(v)), uintptr(unsafe.Pointer(guidKey)), uintptr(value))
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}

// -- IMFSample (inherits IMFAttributes) --
type IMFSample struct { ole.IUnknown }
type IMFSampleVtbl struct {
	ole.IUnknownVtbl
	GetItem             uintptr
	GetItemType         uintptr
	CompareItem         uintptr
	Compare             uintptr
	GetUINT32           uintptr
	GetUINT64           uintptr
	GetDouble           uintptr
	GetGUID             uintptr
	GetStringLength     uintptr
	GetString           uintptr
	GetAllocatedString  uintptr
	GetBlobSize         uintptr
	GetBlob             uintptr
	GetAllocatedBlob    uintptr
	GetUnknown          uintptr
	SetItem             uintptr
	DeleteItem          uintptr
	DeleteAllItems      uintptr
	SetUINT32           uintptr
	SetUINT64           uintptr
	SetDouble           uintptr
	SetGUID             uintptr
	SetString           uintptr
	SetBlob             uintptr
	SetUnknown          uintptr
	LockStore           uintptr
	UnlockStore         uintptr
	GetCount            uintptr
	GetItemByIndex      uintptr
	CopyAllItems        uintptr
	// IMFSample methods follow
	GetSampleFlags      uintptr
	SetSampleFlags      uintptr
	GetSampleTime       uintptr
	SetSampleTime       uintptr
	GetSampleDuration   uintptr
	SetSampleDuration   uintptr
	GetBufferCount      uintptr
	GetBufferByIndex    uintptr
	ConvertToContiguousBuffer uintptr
	AddBuffer           uintptr
	RemoveBufferByIndex uintptr
	RemoveAllBuffers    uintptr
	GetTotalLength      uintptr
	CopyToBuffer        uintptr
}
func (v *IMFSample) VTable() *IMFSampleVtbl { return (*IMFSampleVtbl)(unsafe.Pointer(v.RawVTable)) }
func (v *IMFSample) ConvertToContiguousBuffer(buf **IMFMediaBuffer) error {
	ret, _, _ := syscall.Syscall(v.VTable().ConvertToContiguousBuffer, 2, uintptr(unsafe.Pointer(v)), uintptr(unsafe.Pointer(buf)), 0)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}

// -- IMFMediaBuffer --
type IMFMediaBuffer struct { ole.IUnknown }
type IMFMediaBufferVtbl struct {
	ole.IUnknownVtbl
	Lock             uintptr
	Unlock           uintptr
	GetCurrentLength uintptr
	SetCurrentLength uintptr
	GetMaxLength     uintptr
}
func (v *IMFMediaBuffer) VTable() *IMFMediaBufferVtbl { return (*IMFMediaBufferVtbl)(unsafe.Pointer(v.RawVTable)) }
func (v *IMFMediaBuffer) Lock(ppbBuffer **byte, pcbMaxLength *uint32, pcbCurrentLength *uint32) error {
	ret, _, _ := syscall.Syscall6(v.VTable().Lock, 4, uintptr(unsafe.Pointer(v)), uintptr(unsafe.Pointer(ppbBuffer)), uintptr(unsafe.Pointer(pcbMaxLength)), uintptr(unsafe.Pointer(pcbCurrentLength)), 0, 0)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func (v *IMFMediaBuffer) Unlock() error {
	ret, _, _ := syscall.Syscall(v.VTable().Unlock, 1, uintptr(unsafe.Pointer(v)), 0, 0)
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}

func MFStartup() error {
	ret, _, _ := procMFStartup.Call(uintptr(MF_VERSION), uintptr(MFSTARTUP_LITE))
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func MFShutdown() error {
	ret, _, _ := procMFShutdown.Call()
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func MFCreateMediaType(mt **IMFMediaType) error {
	ret, _, _ := procMFCreateMediaType.Call(uintptr(unsafe.Pointer(mt)))
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
func MFCreateSourceReaderFromURL(url *uint16, attributes uintptr, reader **IMFSourceReader) error {
	ret, _, _ := procMFCreateSourceReaderFromURL.Call(uintptr(unsafe.Pointer(url)), attributes, uintptr(unsafe.Pointer(reader)))
	if ret != 0 { return syscall.Errno(ret) }
	return nil
}
