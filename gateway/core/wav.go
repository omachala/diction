package core

import (
	"encoding/binary"
	"io"
)

// WriteWAVHeader writes a 44-byte RIFF WAV header for 16kHz, 16-bit, mono PCM.
func WriteWAVHeader(w io.Writer, dataSize int) error {
	const (
		sampleRate    = 16000
		bitsPerSample = 16
		numChannels   = 1
		byteRate      = sampleRate * numChannels * bitsPerSample / 8
		blockAlign    = numChannels * bitsPerSample / 8
	)

	fileSize := uint32(36 + dataSize)

	var header [44]byte
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], fileSize)
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16) // fmt chunk size
	binary.LittleEndian.PutUint16(header[20:22], 1)  // PCM
	binary.LittleEndian.PutUint16(header[22:24], numChannels)
	binary.LittleEndian.PutUint32(header[24:28], sampleRate)
	binary.LittleEndian.PutUint32(header[28:32], byteRate)
	binary.LittleEndian.PutUint16(header[32:34], blockAlign)
	binary.LittleEndian.PutUint16(header[34:36], bitsPerSample)
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataSize))

	_, err := w.Write(header[:])
	return err
}
