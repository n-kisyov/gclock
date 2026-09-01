package audio

import (
	"math"
	"sync/atomic"
	"time"
	"unsafe"
	"syscall"

	"github.com/go-ole/go-ole"
	"github.com/lxn/win"
	"github.com/moutend/go-wca/pkg/wca"
	"golang.org/x/sys/windows"
)

const (
	AudioBufferMs = 200
	RefTimesPerMs = 10000

	ToneAHz     = 1000.0
	ToneBHz     = 1200.0
	ToneOnMs    = 200.0
	ToneGapMs   = 80.0
	ToneTailMs  = 500.0
	TonePeriodMs = ToneOnMs + ToneGapMs + ToneOnMs + ToneTailMs
)

var (
	gMfStarted   bool
	gStopChan    chan struct{}
	gNotifyWnd   win.HWND
	gPath        string
	gUseTone     bool
	gPlaying     atomic.Bool

	gGain        float32 = 1.0
	gRampFrom    float32 = 1.0
	gRampTo      float32 = 1.0
	gRampMs      uint32  = 0
	gRampStart   time.Time

	gToneFrame   uint64 = 0

	// Source reader
	gFs struct {
		reader      *IMFSourceReader
		buf         *IMFMediaBuffer
		locked      *byte
		lockedBytes uint32
		offsetBytes uint32
		eof         bool
	}
)

func Init() bool {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if !gMfStarted {
		if err := MFStartup(); err != nil {
			return false
		}
		gMfStarted = true
	}
	return true
}

func Shutdown() {
	Stop()
	if gMfStarted {
		MFShutdown()
		gMfStarted = false
	}
	ole.CoUninitialize()
}

func fileReleaseBuffer() {
	if gFs.buf != nil {
		if gFs.locked != nil {
			gFs.buf.Unlock()
		}
		gFs.buf.Release()
	}
	gFs.buf = nil
	gFs.locked = nil
	gFs.lockedBytes = 0
	gFs.offsetBytes = 0
}

func fileClose() {
	fileReleaseBuffer()
	if gFs.reader != nil {
		gFs.reader.Release()
		gFs.reader = nil
	}
	gFs.eof = false
}

func fileOpen(path string, rate uint32, channels uint32) bool {
	gFs.reader = nil
	gFs.buf = nil
	gFs.locked = nil
	gFs.lockedBytes = 0
	gFs.offsetBytes = 0
	gFs.eof = false

	urlPtr, _ := syscall.UTF16PtrFromString(path)
	var reader *IMFSourceReader
	if err := MFCreateSourceReaderFromURL(urlPtr, 0, &reader); err != nil {
		return false
	}
	gFs.reader = reader

	reader.SetStreamSelection(MF_SOURCE_READER_ALL_STREAMS, false)
	reader.SetStreamSelection(MF_SOURCE_READER_FIRST_AUDIO_STREAM, true)

	var mt *IMFMediaType
	if err := MFCreateMediaType(&mt); err != nil {
		fileClose()
		return false
	}

	mt.SetGUID(&MF_MT_MAJOR_TYPE, &MFMediaType_Audio)
	mt.SetGUID(&MF_MT_SUBTYPE, &MFAudioFormat_Float)
	mt.SetUINT32(&MF_MT_AUDIO_SAMPLES_PER_SECOND, rate)
	mt.SetUINT32(&MF_MT_AUDIO_NUM_CHANNELS, channels)
	mt.SetUINT32(&MF_MT_AUDIO_BITS_PER_SAMPLE, 32)
	mt.SetUINT32(&MF_MT_AUDIO_BLOCK_ALIGNMENT, channels*4)
	mt.SetUINT32(&MF_MT_AUDIO_AVG_BYTES_PER_SECOND, rate*channels*4)
	mt.SetUINT32(&MF_MT_ALL_SAMPLES_INDEPENDENT, 1)

	if err := reader.SetCurrentMediaType(MF_SOURCE_READER_FIRST_AUDIO_STREAM, 0, mt); err != nil {
		mt.Release()
		fileClose()
		return false
	}
	mt.Release()
	return true
}

func fileNextBuffer() bool {
	for attempt := 0; attempt < 8; attempt++ {
		var flags uint32
		var sample *IMFSample
		err := gFs.reader.ReadSample(MF_SOURCE_READER_FIRST_AUDIO_STREAM, 0, nil, &flags, nil, &sample)
		if err != nil {
			gFs.eof = true
			return false
		}
		if (flags & MF_SOURCE_READERF_ENDOFSTREAM) != 0 {
			if sample != nil { sample.Release() }
			gFs.eof = true
			return false
		}
		if sample == nil {
			continue
		}

		err = sample.ConvertToContiguousBuffer(&gFs.buf)
		sample.Release()
		if err != nil || gFs.buf == nil {
			gFs.buf = nil
			continue
		}

		err = gFs.buf.Lock(&gFs.locked, nil, &gFs.lockedBytes)
		if err != nil {
			gFs.buf.Release()
			gFs.buf = nil
			gFs.locked = nil
			continue
		}
		gFs.offsetBytes = 0
		if gFs.lockedBytes == 0 {
			fileReleaseBuffer()
			continue
		}
		return true
	}
	gFs.eof = true
	return false
}

func fileFill(dst []float32, frames uint32, channels uint32) uint32 {
	produced := uint32(0)
	frameBytes := channels * 4

	for produced < frames {
		if gFs.locked == nil {
			if gFs.eof { break }
			if !fileNextBuffer() { break }
		}
		availBytes := gFs.lockedBytes - gFs.offsetBytes
		availFrames := availBytes / frameBytes
		want := frames - produced
		n := availFrames
		if want < n {
			n = want
		}

		if n > 0 {
			// Copy n*channels floats
			srcPtr := unsafe.Pointer(uintptr(unsafe.Pointer(gFs.locked)) + uintptr(gFs.offsetBytes))
			srcSlice := unsafe.Slice((*float32)(srcPtr), n*channels)
			copy(dst[produced*channels:], srcSlice)
			produced += n
			gFs.offsetBytes += n * frameBytes
		}
		if gFs.offsetBytes+frameBytes > gFs.lockedBytes {
			fileReleaseBuffer()
		}
	}
	return produced
}

func toneEnvelope(posMs, lenMs float64) float64 {
	edge := 5.0
	if posMs < edge { return posMs / edge }
	if posMs > lenMs-edge { return (lenMs - posMs) / edge }
	return 1.0
}

func toneFill(dst []float32, frames uint32, channels uint32, rate uint32) uint32 {
	for i := uint32(0); i < frames; i++ {
		tMs := (float64(gToneFrame+uint64(i)) * 1000.0) / float64(rate)
		pos := math.Mod(tMs, TonePeriodMs)

		freq := 0.0
		local := 0.0
		if pos < ToneOnMs {
			freq = ToneAHz
			local = pos
		} else if pos >= ToneOnMs+ToneGapMs && pos < ToneOnMs+ToneGapMs+ToneOnMs {
			freq = ToneBHz
			local = pos - (ToneOnMs + ToneGapMs)
		}

		v := float32(0.0)
		if freq > 0.0 {
			env := toneEnvelope(local, ToneOnMs)
			v = float32(0.45 * env * math.Sin(2.0*math.Pi*freq*(local/1000.0)))
		}
		for c := uint32(0); c < channels; c++ {
			dst[i*channels+c] = v
		}
	}
	gToneFrame += uint64(frames)
	return frames
}

func currentGain() float32 {
	ms := gRampMs
	if ms == 0 {
		return gGain
	}
	elapsed := uint32(time.Since(gRampStart).Milliseconds())
	if elapsed >= ms {
		gGain = gRampTo
		gRampMs = 0
		return gGain
	}
	t := float32(elapsed) / float32(ms)
	g := gRampFrom + (gRampTo-gRampFrom)*t
	gGain = g
	return g
}

func SetGain(gain float32) {
	if gain < 0.0 { gain = 0.0 }
	if gain > 1.0 { gain = 1.0 }
	gRampMs = 0
	gGain = gain
}

func GetGain() float32 {
	return gGain
}

func RampGain(from, to float32, ms uint32) {
	if from < 0.0 { from = 0.0 }
	if from > 1.0 { from = 1.0 }
	if to < 0.0 { to = 0.0 }
	if to > 1.0 { to = 1.0 }

	if ms == 0 {
		SetGain(to)
		return
	}
	gGain = from
	gRampFrom = from
	gRampTo = to
	gRampStart = time.Now()
	gRampMs = ms
}

func RampActive() bool {
	return gRampMs != 0
}

func formatIsFloat(wf *wca.WAVEFORMATEX) bool {
	if wf.WFormatTag == 3 { return true } // WAVE_FORMAT_IEEE_FLOAT
	if wf.WFormatTag == 0xFFFE { // WAVE_FORMAT_EXTENSIBLE
		// The GUID for WAVE_SUBFORMAT_IEEE_FLOAT is: 0x00000003, 0x0000, 0x0010, 0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71
		// It's located right after WAVEFORMATEX. We can just read the first uint32.
		data1 := *(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(wf)) + unsafe.Sizeof(*wf) + 2))
		return data1 == 0x00000003
	}
	return false
}

func writeFrames(out []byte, src []float32, frames uint32, channels uint32, isFloat bool, bytesPerSample uint32, gain float32) {
	samples := frames * channels
	if isFloat {
		dst := unsafe.Slice((*float32)(unsafe.Pointer(&out[0])), samples)
		for i := uint32(0); i < samples; i++ {
			dst[i] = src[i] * gain
		}
		return
	}
	if bytesPerSample == 2 {
		dst := unsafe.Slice((*int16)(unsafe.Pointer(&out[0])), samples)
		for i := uint32(0); i < samples; i++ {
			v := src[i] * gain
			if v > 1.0 { v = 1.0 }
			if v < -1.0 { v = -1.0 }
			dst[i] = int16(v * 32767.0)
		}
		return
	}
	for i := range out {
		out[i] = 0
	}
}

func renderThread(stopChan chan struct{}, readyChan chan bool, useTone bool, path string, notifyWnd win.HWND) {
	ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	defer ole.CoUninitialize()

	var enumr *wca.IMMDeviceEnumerator
	var dev *wca.IMMDevice
	var client *wca.IAudioClient
	var render *wca.IAudioRenderClient
	var mix *wca.WAVEFORMATEX

	cleanup := func() {
		if render != nil { render.Release() }
		if client != nil { client.Release() }
		if mix != nil { ole.CoTaskMemFree(uintptr(unsafe.Pointer(mix))) }
		if dev != nil { dev.Release() }
		if enumr != nil { enumr.Release() }
		fileClose()
	}
	defer cleanup()

	wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &enumr)
	if enumr != nil {
		enumr.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &dev)
	}
	if dev != nil {
		dev.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &client)
	}
	if client != nil {
		client.GetMixFormat(&mix)
	}

	started := false
	var bufferEvent windows.Handle
	var bufferFrames uint32
	var channels, rate, bytesPerSample uint32
	var isFloat bool
	var scratch []float32

	if mix != nil {
		dur := int64(AudioBufferMs * RefTimesPerMs)
		if err := client.Initialize(wca.AUDCLNT_SHAREMODE_SHARED, wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK, wca.REFERENCE_TIME(dur), 0, mix, nil); err == nil {
			bufferEvent, _ = windows.CreateEvent(nil, 0, 0, nil)
			if bufferEvent != 0 {
				client.SetEventHandle(uintptr(bufferEvent))
				client.GetBufferSize(&bufferFrames)
				client.GetService(wca.IID_IAudioRenderClient, &render)

				channels = uint32(mix.NChannels)
				rate = mix.NSamplesPerSec
				isFloat = formatIsFloat(mix)
				bytesPerSample = uint32(mix.WBitsPerSample / 8)

				if useTone || fileOpen(path, rate, channels) {
					scratch = make([]float32, bufferFrames*channels)
					started = true
				}
			}
		}
	}

	readyChan <- started
	if !started {
		if bufferEvent != 0 { windows.CloseHandle(bufferEvent) }
		return
	}
	defer windows.CloseHandle(bufferEvent)

	gToneFrame = 0
	client.Start()
	defer client.Stop()

	reachedEnd := false

	for {
		select {
		case <-stopChan:
			goto WaitDrain
		default:
		}

		w, _ := windows.WaitForSingleObject(bufferEvent, 2000)
		if w != windows.WAIT_OBJECT_0 && w != 258 { // 258 is WAIT_TIMEOUT
			break
		}

		select {
		case <-stopChan:
			goto WaitDrain
		default:
		}

		var padding uint32
		if err := client.GetCurrentPadding(&padding); err != nil { break }

		avail := bufferFrames - padding
		if avail == 0 { continue }

		var out *byte
		if err := render.GetBuffer(avail, &out); err != nil { break }

		got := uint32(0)
		if useTone {
			got = toneFill(scratch, avail, channels, rate)
		} else {
			got = fileFill(scratch, avail, channels)
		}

		if got == 0 {
			render.ReleaseBuffer(0, wca.AUDCLNT_BUFFERFLAGS_SILENT)
			reachedEnd = true
			break
		}

		outSlice := unsafe.Slice(out, got*channels*bytesPerSample)
		writeFrames(outSlice, scratch, got, channels, isFloat, bytesPerSample, currentGain())
		render.ReleaseBuffer(got, 0)
	}

WaitDrain:
	if reachedEnd {
		var padding uint32
		for i := 0; i < 100; i++ {
			select {
			case <-stopChan:
				reachedEnd = false
				goto Done
			default:
			}
			if err := client.GetCurrentPadding(&padding); err != nil || padding == 0 {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
Done:

	gPlaying.Store(false)

	if reachedEnd && notifyWnd != 0 {
		win.PostMessage(notifyWnd, win.WM_APP+3, 0, 0) // WM_AUDIO_TRACK_DONE
	}
}

func startInternal(tone bool, path string, notifyWnd win.HWND) bool {
	Stop()
	gUseTone = tone
	gPath = path

	stopChan := make(chan struct{})
	readyChan := make(chan bool, 1)

	gPlaying.Store(true)
	gStopChan = stopChan

	go renderThread(stopChan, readyChan, tone, path, notifyWnd)

	select {
	case ok := <-readyChan:
		if !ok {
			gPlaying.Store(false)
			return false
		}
		return true
	case <-time.After(4 * time.Second):
		Stop()
		return false
	}
}

func PlayFile(path string, notifyWnd win.HWND) bool {
	if path == "" { return false }
	return startInternal(false, path, notifyWnd)
}

func PlayTone(notifyWnd win.HWND) bool {
	return startInternal(true, "", notifyWnd)
}

func Stop() {
	if gStopChan != nil {
		close(gStopChan)
		gStopChan = nil
	}
	// wait for thread to actually die
	for i := 0; i < 30; i++ {
		if !gPlaying.Load() { break }
		time.Sleep(100 * time.Millisecond)
	}
	gPlaying.Store(false)
	gRampMs = 0
}

func IsPlaying() bool {
	return gPlaying.Load()
}
