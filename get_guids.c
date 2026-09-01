#include <windows.h>
#include <mfapi.h>
#include <stdio.h>
#include <initguid.h>
#include <mfidl.h>
#include <mfreadwrite.h>

void print_guid(const char *name, const GUID *g) {
    printf("var %s = windows.GUID{Data1: 0x%08lX, Data2: 0x%04X, Data3: 0x%04X, Data4: [8]byte{0x%02X, 0x%02X, 0x%02X, 0x%02X, 0x%02X, 0x%02X, 0x%02X, 0x%02X}}\n",
        name, g->Data1, g->Data2, g->Data3,
        g->Data4[0], g->Data4[1], g->Data4[2], g->Data4[3],
        g->Data4[4], g->Data4[5], g->Data4[6], g->Data4[7]);
}

int main() {
    print_guid("MF_MT_MAJOR_TYPE", &MF_MT_MAJOR_TYPE);
    print_guid("MFMediaType_Audio", &MFMediaType_Audio);
    print_guid("MF_MT_SUBTYPE", &MF_MT_SUBTYPE);
    print_guid("MFAudioFormat_Float", &MFAudioFormat_Float);
    print_guid("MF_MT_AUDIO_SAMPLES_PER_SECOND", &MF_MT_AUDIO_SAMPLES_PER_SECOND);
    print_guid("MF_MT_AUDIO_NUM_CHANNELS", &MF_MT_AUDIO_NUM_CHANNELS);
    print_guid("MF_MT_AUDIO_BITS_PER_SAMPLE", &MF_MT_AUDIO_BITS_PER_SAMPLE);
    print_guid("MF_MT_AUDIO_BLOCK_ALIGNMENT", &MF_MT_AUDIO_BLOCK_ALIGNMENT);
    print_guid("MF_MT_AUDIO_AVG_BYTES_PER_SECOND", &MF_MT_AUDIO_AVG_BYTES_PER_SECOND);
    print_guid("MF_MT_ALL_SAMPLES_INDEPENDENT", &MF_MT_ALL_SAMPLES_INDEPENDENT);
    
    printf("MF_SOURCE_READER_ALL_STREAMS = %lu\n", (unsigned long)MF_SOURCE_READER_ALL_STREAMS);
    printf("MF_SOURCE_READER_FIRST_AUDIO_STREAM = %lu\n", (unsigned long)MF_SOURCE_READER_FIRST_AUDIO_STREAM);
    printf("MF_SOURCE_READERF_ENDOFSTREAM = %lu\n", (unsigned long)MF_SOURCE_READERF_ENDOFSTREAM);
    return 0;
}
